package root

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing/eventstore"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/consumerpages"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/customerstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// The values are invented, with values that a span has no other reason to contain. The comparison has no difference
// for upper and lower case.
var identificationValues = []string{
	"00063927", "63927", "kangaroo.paw@mail.test", "kangaroo.paw", "wallaroo.gone@mail.test", "wallaroo.gone",
}

// failingMail fails for the address of the customer with an error that contains the address, as an SMTP error can.
type failingMail struct {
	failFor string
	sent    int
}

func (m *failingMail) Send(_ context.Context, to string, _ domain.Email) error {
	if strings.EqualFold(to, m.failFor) {
		return errors.New("550 mailbox " + to + " is not available")
	}

	m.sent++

	return nil
}

// identificationFlow is one browser in one flow of the consumer pages, behind the server telemetry of the data holder.
type identificationFlow struct {
	t       *testing.T
	handler http.Handler
	cookies []*http.Cookie
}

var formTokenField = regexp.MustCompile(`name="csrf_token" value="([^"]+)"`)

func newIdentificationHandler(pages consumerpages.Pages) http.Handler {
	states := map[string]any{}
	mux := http.NewServeMux()
	mux.HandleFunc("/authorize/{flow}", func(w http.ResponseWriter, r *http.Request) {
		flow := r.PathValue("flow")
		outcome := pages.Interact(w, r, authserver.Interaction{
			ClientID: "client-1", Scopes: []string{"openid"}, Path: r.URL.Path, State: states[flow],
		})

		if next, ok := outcome.Continues(); ok {
			states[flow] = next
		}
	})

	return telemetry.NewHTTPHandler(mux)
}

func (f *identificationFlow) do(method, flow string, form url.Values) *httptest.ResponseRecorder {
	f.t.Helper()

	req := httptest.NewRequestWithContext(
		f.t.Context(), method, "https://bank.test/authorize/"+flow, strings.NewReader(form.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for _, cookie := range f.cookies {
		req.AddCookie(cookie)
	}

	rec := httptest.NewRecorder()
	f.handler.ServeHTTP(rec, req)

	if cookies := rec.Result().Cookies(); len(cookies) > 0 {
		f.cookies = cookies
	}

	return rec
}

// identify gives the identifier on the first page of a new flow, and gives the form token of the flow. A flow that
// ends gives no page, because the authorisation server writes the redirect.
func (f *identificationFlow) identify(flow, identifier string) string {
	f.t.Helper()

	first := f.do(http.MethodGet, flow, nil)
	require.Equal(f.t, http.StatusOK, first.Code)

	token := formTokenField.FindStringSubmatch(first.Body.String())[1]
	otpPage := f.do(http.MethodPost, flow, url.Values{
		"csrf_token": {token}, "action": {"continue"}, "identifier": {identifier},
	})
	require.Equal(f.t, http.StatusOK, otpPage.Code, otpPage.Body.String())

	return token
}

func (f *identificationFlow) resend(flow, token string) {
	f.t.Helper()

	resent := f.do(http.MethodPost, flow, url.Values{"csrf_token": {token}, "action": {"resend"}})
	require.Equal(f.t, http.StatusOK, resent.Code)
}

// Verifies: DH-0008
func TestTheIdentificationOfTheConsumerAuthenticationPutsNoEmailAddressInTheSpansMetricsOrLogs(t *testing.T) {
	spans, metrics, logs := recordTelemetry(t)

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	store := customerstore.New(pool, nil, nil)
	require.NoError(t, store.AddCustomer(t.Context(), domain.Customer{
		CustomerNumber: "00063927", LastName: "Quoll", EmailAddress: "Kangaroo.Paw@mail.test", LastUpdated: time.Now(),
	}))

	events := eventstore.New(pool, registrationstore.Projections()...)
	require.NoError(t, app.ClientRegistrationType.Execute(t.Context(), events, "software-1", domain.RegisterClient{
		Registration: domain.Registration{
			ClientID: "client-1", SoftwareID: "software-1", ClientName: "Recipient Web", OrgName: "Recipient",
		},
	}))

	mail := &failingMail{}
	pages := consumerpages.New("Coastal Bank", registrationstore.NewQueries(pool), app.Authenticator{
		Customers: store, Mail: mail, Random: func() uint64 { return 424_242 }, Now: time.Now,
	}, app.Consents{}, time.Now)
	browser := &identificationFlow{t: t, handler: newIdentificationHandler(pages)}
	setUpSpans := len(spans.Ended())

	for flow, identifier := range map[string]string{
		"flow-known": " KANGAROO.PAW@Mail.test ", "flow-unknown": "Wallaroo.Gone@mail.test", "flow-number": "00063927",
	} {
		browser.resend(flow, browser.identify(flow, identifier))
	}

	require.Equal(t, 4, mail.sent, "the known address and the customer number each get two OTPs")

	mail.failFor = "kangaroo.paw@mail.test"

	browser.identify("flow-mail-failure", "kangaroo.paw@mail.test")

	ended := spans.Ended()[setUpSpans:]
	require.NotEmpty(t, ended)

	for _, span := range ended {
		text := spanText(span)
		for _, value := range identificationValues {
			assert.NotContains(t, text, value, "span %q", span.Name())
		}
	}

	text := metricText(t, metrics)
	assert.NotContains(
		t,
		text,
		"/authorize/{flow}",
		"the requests make no metrics, because their spans give the detail",
	)
	assert.Contains(t, logs.String(), "authorisation flow failed", "the test sees the log of the mail failure")

	for _, value := range identificationValues {
		assert.NotContains(t, text, value)
		assert.NotContains(t, strings.ToLower(logs.String()), value)
	}
}
