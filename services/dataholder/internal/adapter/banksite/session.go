package banksite

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// Customers gives domain.ValidationError and app.ErrEmailAddressUsed from Create, and app.ErrCustomerNotFound from
// Customer and Identify.
// CloseAccountThatEnds gives app.ErrEndsChanged, app.ErrCustomerNotFound, app.ErrAccountNotFound and
// domain.ErrAccountClosed.
type Customers interface {
	Create(ctx context.Context, customer domain.Customer) (domain.Customer, error)
	Customer(ctx context.Context, customerNumber string) (app.CustomerRecord, error)
	Identify(ctx context.Context, identifier string) (app.CustomerRecord, error)
	CloseAccountThatEnds(ctx context.Context, customerNumber, accountID string, ends []string) (domain.Account, error)
}

const (
	// A sign-up, a sign-in or an open is a few hundred bytes.
	maxBodyBytes = 4 << 10
	notSignedIn  = "We could not sign you in. Check your email address or customer number."
	emailInUse   = "A customer of the bank has this email address. Sign in, or enter a different email address."
)

// sessions joins the store to the session cookie of the brand. Cookies are not isolated by port, so the name has the
// brand ID: a different bank on the same host has a cookie of its own.
type sessions struct {
	store      *sessionStore
	cookieName string
}

func newSessions(store *sessionStore, brandID string) sessions {
	return sessions{store: store, cookieName: "__Host-bank-session-" + brandID}
}

// customerOf gives the customer number of the valid session of the request. A handler of a page with the data of a
// customer is behind require, and gets the customer with customerOf.
func (s sessions) customerOf(r *http.Request) (customerNumber string, ok bool) {
	cookie, err := r.Cookie(s.cookieName)
	if err != nil {
		return "", false
	}

	return s.store.customer(cookie.Value)
}

// require gives 403 to a request with no valid session. With a session cookie, the problem has a type of its own, so
// that the UI can say that the bank signed the consumer out, and the response expires the cookie.
func (s sessions) require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := s.customerOf(r); !ok {
			s.refuse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// refuse is the response of require, for a handler whose session ended after require.
func (s sessions) refuse(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie(s.cookieName); err != nil {
		problem.Write(w, http.StatusForbidden, "Sign in to online banking to see this page.")
		return
	}

	s.expire(w)
	problem.WriteSessionEnded(w, http.StatusForbidden, "The bank signed you out. Sign in again.")
}

// start ends the session of the request, so that each sign-in and sign-up has a new session ID.
func (s sessions) start(w http.ResponseWriter, r *http.Request, place *place, customerNumber string) {
	s.endOf(r)
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    place.start(customerNumber),
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s sessions) endOf(r *http.Request) {
	if cookie, err := r.Cookie(s.cookieName); err == nil {
		s.store.end(cookie.Value)
	}
}

func (s sessions) expire(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// reserve writes 503 when the store is full.
func (s sessions) reserve(w http.ResponseWriter) (*place, bool) {
	place, err := s.store.reserve()
	if err != nil {
		writeFull(w)
		return nil, false
	}

	return place, true
}

// reserveForSignIn uses the place of a valid session of the request when the store is full, because the sign-in ends
// that session.
func (s sessions) reserveForSignIn(w http.ResponseWriter, r *http.Request) (*place, bool) {
	place, err := s.store.reserve()

	_, signedIn := s.customerOf(r)
	if errors.Is(err, errNoPlace) && signedIn {
		s.endOf(r)
		place, err = s.store.reserve()
	}

	if err != nil {
		writeFull(w)
		return nil, false
	}

	return place, true
}

func writeFull(w http.ResponseWriter) {
	problem.Write(w, http.StatusServiceUnavailable, "The bank cannot sign you in now. Try again later.")
}

type customerCreate struct {
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
}

type sessionCustomer struct {
	CustomerNumber string `json:"customerNumber"`
	FirstName      string `json:"firstName,omitempty"`
	LastName       string `json:"lastName"`
}

func sessionCustomerOf(c domain.Customer) sessionCustomer {
	return sessionCustomer{CustomerNumber: c.CustomerNumber, FirstName: c.FirstName, LastName: c.LastName}
}

func signUp(customers Customers, s sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A pointer is nil after the JSON null, which decodes into a struct with no error.
		var request *customerCreate
		if jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxBodyBytes), &request) != nil || request == nil {
			problem.Write(w, http.StatusBadRequest,
				"The request body must be a JSON object with only firstName, lastName and emailAddress.")

			return
		}

		place, ok := s.reserve(w)
		if !ok {
			return
		}
		defer place.release()

		created, err := customers.Create(r.Context(), domain.Customer{
			FirstName: request.FirstName, LastName: request.LastName, EmailAddress: request.EmailAddress,
		})
		if err != nil {
			writeSignUpError(w, r, err)
			return
		}

		s.start(w, r, place, created.CustomerNumber)
		jsonhttp.Write(w, http.StatusCreated, sessionCustomerOf(created))
	})
}

// fieldTexts are for a consumer, who reads the labels of the page and not the names of the fields.
var fieldTexts = map[string]string{
	"firstName": fmt.Sprintf(
		"Enter a first name of %d characters or fewer, or leave the first name empty.", domain.MaxNameCharacters,
	),
	"lastName":     fmt.Sprintf("Enter a last name of 1 to %d characters.", domain.MaxNameCharacters),
	"emailAddress": "Enter an email address, for example name@example.test.",
}

func writeSignUpError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, app.ErrEmailAddressUsed) {
		problem.WriteFieldErrors(w, http.StatusConflict, "", []problem.FieldError{{
			Pointer: problem.FieldPointer("emailAddress"), Detail: emailInUse,
		}})

		return
	}

	var invalid domain.ValidationError
	if !errors.As(err, &invalid) {
		problem.WriteUnexpected(w, r, err, notCompleted)
		return
	}

	errs := make([]problem.FieldError, 0, len(invalid.Errors))
	for _, f := range invalid.Errors {
		text, ok := fieldTexts[f.Field]
		if !ok {
			text = "Check this value."
		}

		errs = append(errs, problem.FieldError{Pointer: problem.FieldPointer(f.Field), Detail: text})
	}

	problem.WriteFieldErrors(w, http.StatusBadRequest, "", errs)
}

// signInRequest has no password: the UI does not send it, and a body with a password field is not valid.
type signInRequest struct {
	// Identifier is the email address or the customer number.
	Identifier string `json:"identifier"`
}

func signIn(customers Customers, s sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request *signInRequest
		if jsonhttp.Decode(http.MaxBytesReader(w, r.Body, maxBodyBytes), &request) != nil || request == nil {
			problem.Write(w, http.StatusBadRequest, "The request body must be a JSON object with only identifier.")
			return
		}

		place, ok := s.reserveForSignIn(w, r)
		if !ok {
			return
		}
		defer place.release()

		record, err := customers.Identify(r.Context(), request.Identifier)
		if errors.Is(err, app.ErrCustomerNotFound) {
			problem.WriteFieldErrors(w, http.StatusBadRequest, notSignedIn, []problem.FieldError{{
				Pointer: problem.FieldPointer("identifier"), Detail: notSignedIn,
			}})

			return
		}

		if err != nil {
			problem.WriteUnexpected(w, r, err, notCompleted)
			return
		}

		s.start(w, r, place, record.Customer.CustomerNumber)
		w.WriteHeader(http.StatusNoContent)
	})
}

func readSession(customers Customers, s sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		customerNumber, ok := s.customerOf(r)
		if !ok {
			s.refuse(w, r)
			return
		}

		record, err := customers.Customer(r.Context(), customerNumber)
		if err != nil {
			writeAccountError(w, r, s, err)
			return
		}

		jsonhttp.Write(w, http.StatusOK, sessionCustomerOf(record.Customer))
	})
}

// signOut needs no session: the consumer is signed out after it in each case.
func signOut(s sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.endOf(r)
		s.expire(w)
		w.WriteHeader(http.StatusNoContent)
	})
}
