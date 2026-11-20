// Package consumerpages serves the pages on which a consumer authenticates and gives consent in an authorisation flow.
package consumerpages

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// The __Host- prefix makes a browser accept the cookie only for this host, over TLS and with the path "/".
const (
	cookieName     = "__Host-cdr-flow"
	tokenField     = "csrf_token"
	actionField    = "action"
	actionContinue = "continue"
	actionResend   = "resend"
	actionCancel   = "cancel"
	actionConfirm  = "confirm"
	accountField   = "account"
)

//go:embed templates/*.html
var templateFiles embed.FS

var templates = template.Must(
	template.New("pages").
		Funcs(template.FuncMap{"initial": initial, "demonstration": domain.DemonstrationNotice}).
		ParseFS(templateFiles, "templates/*.html"),
)

func initial(name string) string {
	first, _ := utf8.DecodeRuneInString(name)

	return string(unicode.ToUpper(first))
}

type step string

const (
	stepCustomerNumber step = "customer-number"
	stepOTP            step = "otp"
	stepConsent        step = "consent"
)

// flow is the state of one authorisation flow between two requests.
type flow struct {
	step step
	// token is in the cookie of the browser that started the flow and in each form of the flow.
	token string
	// identifier is the entry of the consumer: an email address or a customer number. It is only for a new OTP. After
	// the authentication, auth gives the customer.
	identifier string
	auth       domain.Authentication
	// authenticatedAt is the time of the correct OTP entry.
	authenticatedAt time.Time
}

// Authenticator has the contract of app.Authenticator.
type Authenticator interface {
	SendOTP(
		ctx context.Context,
		auth domain.Authentication,
		identifier, recipientBrand string,
	) (domain.Authentication, error)
	Verify(auth domain.Authentication, entry string) (domain.Authentication, domain.OTPResult)
}

// Consenter has the contract of app.Consents.
type Consenter interface {
	Accounts(ctx context.Context, customerNumber string) ([]domain.Account, error)
	Confirm(ctx context.Context, request domain.ConsentRequest, selected []string) (app.Confirmation, error)
}

type Pages struct {
	// bank is the brand name of the data holder.
	bank          string
	registrations app.RegistrationReader
	authenticator Authenticator
	consents      Consenter
	now           func() time.Time
}

var _ authserver.ConsumerPages = Pages{}

func New(
	bank string,
	registrations app.RegistrationReader,
	authenticator Authenticator,
	consents Consenter,
	now func() time.Time,
) Pages {
	return Pages{bank: bank, registrations: registrations, authenticator: authenticator, consents: consents, now: now}
}

type view struct {
	// Bank is the brand name of the data holder, and Recipient is the brand name of the data recipient.
	Bank      string
	Recipient string
	Title     string
	// Step is the step of the page in the flow, or 0 for a page that is not a step.
	Step   int
	Action string
	Token  string
	// Notice tells the consumer the result of the last form.
	Notice string
	// CanResend is false when the flow sent its last OTP.
	CanResend bool
	// ValidMinutes is the time for which an OTP is valid.
	ValidMinutes int
	Consent      consentView
}

// consentView has the request of the software product in the language that the CDR Rules and the CDS require.
type consentView struct {
	// LegalEntity is the name of the accredited person, and Software is the name of its software product.
	LegalEntity string
	Software    string
	Clusters    []domain.DataCluster
	// Eligible is false when the customer has no open account, and then the consumer can only cancel.
	Eligible      bool
	NeedsAccounts bool
	Accounts      []accountView
	// ClosedAccounts are on the page with no control, because the data holder does not share their data.
	ClosedAccounts []accountView
	OneOccasion    bool
	// Duration and EndDate are for a consent that is not for one occasion. The EndDate is in UTC.
	Duration string
	EndDate  string
}

// Page is for a template, which gives its title and its step.
func (v view) Page(title string, step int) view {
	v.Title, v.Step = title, step

	return v
}

type accountView struct {
	ID           string
	DisplayName  string
	MaskedNumber string
}

const (
	noticeIncorrect  = "That one-time password is not correct or is no longer valid. Try again."
	noticeResent     = "We sent you a new one-time password. The one before it is no longer valid."
	noticeNoAccount  = "Select one account or more to share your data."
	noticeNotAccount = "%s cannot share an account that you selected. Select your accounts again."
)

const (
	hoursInDay    = 24
	daysInYear    = 365
	endDateLayout = "2 January 2006"
)

// consentStep has the values of one request at the OTP step or the consent step.
type consentStep struct {
	in           authserver.Interaction
	registration domain.Registration
	current      flow
	page         view
}

func (p Pages) Interact(w http.ResponseWriter, r *http.Request, in authserver.Interaction) authserver.Outcome {
	registration, err := p.registrations.Registration(r.Context(), in.ClientID)
	if err != nil {
		return unavailable(r.Context(), err)
	}

	current, started := in.State.(flow)
	if !started {
		current = flow{step: stepCustomerNumber, token: browserToken(w, r)}
	}

	page := view{Bank: p.bank, Recipient: registration.OrgName, Action: in.Path, Token: current.token}
	step := consentStep{in: in, registration: registration, current: current, page: page}

	if r.Method != http.MethodPost {
		switch current.step {
		case stepOTP:
			return p.showOTP(w, http.StatusOK, current, page, "")
		case stepConsent:
			return p.showConsent(w, r, http.StatusOK, step, "")
		case stepCustomerNumber:
		}

		render(w, http.StatusOK, string(current.step), page)

		return authserver.Continue(current)
	}

	if !sameBrowser(r, current.token) {
		page.Token = ""
		render(w, http.StatusForbidden, "form-refused", page)

		return authserver.Continue(current)
	}

	if r.PostFormValue(actionField) == actionCancel {
		return authserver.Deny(authserver.AccessDenied)
	}

	switch current.step {
	case stepOTP:
		return p.verify(w, r, step)
	case stepConsent:
		return p.confirm(w, r, step)
	case stepCustomerNumber:
	}

	return p.identify(w, r, current, page)
}

func unavailable(ctx context.Context, err error) authserver.Outcome {
	// The message of err is not logged, because it can contain query arguments or a mail address.
	slog.ErrorContext(ctx, "authorisation flow failed", "error.type", fmt.Sprintf("%T", err))

	return authserver.Deny(authserver.TemporarilyUnavailable)
}

// identify takes the email address or the customer number, and sends the first OTP.
func (p Pages) identify(w http.ResponseWriter, r *http.Request, current flow, page view) authserver.Outcome {
	identifier := strings.TrimSpace(r.PostFormValue("identifier"))
	if r.PostFormValue(actionField) != actionContinue || identifier == "" {
		render(w, http.StatusBadRequest, string(current.step), page)

		return authserver.Continue(current)
	}

	current.identifier = identifier

	return p.sendOTP(w, r, current, page, "")
}

func (p Pages) verify(w http.ResponseWriter, r *http.Request, step consentStep) authserver.Outcome {
	current, page := step.current, step.page

	switch r.PostFormValue(actionField) {
	case actionResend:
		return p.sendOTP(w, r, current, page, noticeResent)
	case actionContinue:
	default:
		return p.showOTP(w, http.StatusBadRequest, current, page, "")
	}

	auth, result := p.authenticator.Verify(current.auth, r.PostFormValue("otp"))
	current.auth = auth

	switch result {
	case domain.OTPCorrect:
		step.current.auth, step.current.step = auth, stepConsent
		step.current.authenticatedAt = p.now()

		return p.showConsent(w, r, http.StatusOK, step, "")
	case domain.OTPAttemptsUsed:
		return authserver.Deny(authserver.AccessDenied)
	case domain.OTPIncorrect:
	}

	return p.showOTP(w, http.StatusOK, current, page, noticeIncorrect)
}

// sendOTP gives the OTP page also for an identifier that no customer has, so that the page does not show which email
// addresses and customer numbers exist.
func (p Pages) sendOTP(
	w http.ResponseWriter,
	r *http.Request,
	current flow,
	page view,
	notice string,
) authserver.Outcome {
	auth, err := p.authenticator.SendOTP(r.Context(), current.auth, current.identifier, page.Recipient)

	switch {
	case errors.Is(err, domain.ErrTooManyOTPs):
		notice = ""
	case err != nil:
		return unavailable(r.Context(), err)
	}

	current.step, current.auth = stepOTP, auth

	return p.showOTP(w, http.StatusOK, current, page, notice)
}

func (Pages) showOTP(w http.ResponseWriter, status int, current flow, page view, notice string) authserver.Outcome {
	page.Notice = notice
	page.CanResend = current.auth.CanIssue()
	page.ValidMinutes = int(domain.OTPLifetime.Minutes())
	render(w, status, string(stepOTP), page)

	return authserver.Continue(current)
}

func (p Pages) showConsent(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	step consentStep,
	notice string,
) authserver.Outcome {
	in, registration := step.in, step.registration
	consent := consentView{
		LegalEntity:   registration.LegalEntityName,
		Software:      registration.ClientName,
		Clusters:      domain.DataClusters(in.Scopes),
		NeedsAccounts: domain.NeedsAccounts(in.Scopes),
		OneOccasion:   domain.IsOneOccasion(in.SharingDuration),
		Duration:      durationText(in.SharingDuration),
		EndDate:       p.now().UTC().Add(in.SharingDuration).Format(endDateLayout),
	}
	// A registration from an SSA with no legal entity name has only the brand name.
	if consent.LegalEntity == "" {
		consent.LegalEntity = registration.OrgName
	}

	accounts, err := p.consents.Accounts(r.Context(), step.current.auth.CustomerNumber())
	if err != nil {
		return unavailable(r.Context(), err)
	}

	consent.Eligible = len(domain.OpenAccounts(accounts)) > 0

	if consent.NeedsAccounts {
		for _, account := range accounts {
			view := accountView{
				ID:           account.AccountID,
				DisplayName:  account.DisplayName,
				MaskedNumber: account.MaskedNumber(),
			}
			if account.OpenStatus == domain.OpenStatusOpen {
				consent.Accounts = append(consent.Accounts, view)
			} else {
				consent.ClosedAccounts = append(consent.ClosedAccounts, view)
			}
		}
	}

	step.page.Notice, step.page.Consent = notice, consent
	render(w, status, string(stepConsent), step.page)

	return authserver.Continue(step.current)
}

func (p Pages) confirm(w http.ResponseWriter, r *http.Request, step consentStep) authserver.Outcome {
	if r.PostFormValue(actionField) != actionConfirm {
		return p.showConsent(w, r, http.StatusBadRequest, step, "")
	}

	confirmation, err := p.consents.Confirm(r.Context(), domain.ConsentRequest{
		CustomerNumber:   step.current.auth.CustomerNumber(),
		ClientID:         step.in.ClientID,
		SoftwareID:       step.registration.SoftwareID,
		ClientName:       step.registration.ClientName,
		OrgName:          step.registration.OrgName,
		LegalEntityName:  step.registration.LegalEntityName,
		RecipientBaseURI: step.registration.RecipientBaseURI,
		Scopes:           step.in.Scopes,
		SharingDuration:  step.in.SharingDuration,
		SectorIdentifier: app.SectorIdentifier(step.registration),
	}, r.PostForm[accountField])

	switch {
	case errors.Is(err, domain.ErrConsumerNotEligible):
		return p.showConsent(w, r, http.StatusBadRequest, step, "")
	case errors.Is(err, domain.ErrNoAccountSelected):
		return p.showConsent(w, r, http.StatusBadRequest, step, noticeNoAccount)
	case errors.Is(err, domain.ErrAccountNotOfCustomer):
		return p.showConsent(w, r, http.StatusBadRequest, step, fmt.Sprintf(noticeNotAccount, p.bank))
	case err != nil:
		return unavailable(r.Context(), err)
	}

	return authserver.Approve(authserver.Consent{
		ArrangementID:   confirmation.Arrangement.ArrangementID,
		Subject:         confirmation.Subject,
		AuthenticatedAt: step.current.authenticatedAt,
		SharingEndsAt:   confirmation.Arrangement.SharingEndsAt(),
	})
}

// durationText counts a part of an hour as one hour, and gives the longest duration of the CDS as "1 year".
func durationText(sharing time.Duration) string {
	totalHours := int((sharing + time.Hour - 1) / time.Hour)
	days, hours := totalHours/hoursInDay, totalHours%hoursInDay

	if days == daysInYear && hours == 0 {
		return "1 year"
	}

	var parts []string
	if days > 0 {
		parts = append(parts, quantity(days, "day"))
	}

	if hours > 0 {
		parts = append(parts, quantity(hours, "hour"))
	}

	return strings.Join(parts, " and ")
}

func quantity(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}

	return fmt.Sprintf("%d %ss", n, unit)
}

func (p Pages) Refused(w http.ResponseWriter, _ *http.Request, status int) {
	render(w, status, "refused", view{Bank: p.bank})
}

// browserToken gives the token of the cookie that the browser has, or sets a cookie with a new token.
func browserToken(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(cookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	token := rand.Text() + rand.Text()

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return token
}

// sameBrowser is true when the form comes from a page of this flow in the browser that started the flow.
func sameBrowser(r *http.Request, token string) bool {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return false
	}

	fromForm := subtle.ConstantTimeCompare([]byte(r.PostFormValue(tokenField)), []byte(token)) == 1
	fromBrowser := subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(token)) == 1

	return fromForm && fromBrowser
}

func render(w http.ResponseWriter, status int, name string, page view) {
	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")
	header.Set("Cache-Control", "no-store")
	// No form-action: a browser applies it also to the redirect to the data recipient after a form.
	header.Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; font-src 'self'; img-src 'self'; "+
		"frame-ancestors 'none'; base-uri 'none'")
	// The path of a page identifies its flow, so the browser must not send it to the software product.
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	if err := templates.ExecuteTemplate(w, name+".html", page); err != nil {
		panic(err)
	}
}
