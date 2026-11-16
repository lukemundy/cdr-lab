package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const knownCustomer = "00000001"

type sentEmail struct {
	to    string
	email domain.Email
}

type outbox struct {
	sent []sentEmail
	err  error
}

func (o *outbox) Send(_ context.Context, to string, email domain.Email) error {
	if o.err != nil {
		return o.err
	}

	o.sent = append(o.sent, sentEmail{to: to, email: email})

	return nil
}

type authSetup struct {
	authenticator app.Authenticator
	store         *fakeCustomerStore
	mail          *outbox
	now           time.Time
}

func newAuthSetup() *authSetup {
	s := &authSetup{
		store: newFakeCustomerStore(),
		mail:  &outbox{},
		now:   time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC),
	}
	s.store.customers[knownCustomer] = domain.Customer{
		CustomerNumber: knownCustomer, LastName: "Example", EmailAddress: "alex@example.com",
	}
	s.authenticator = app.Authenticator{
		Customers: s.store,
		Mail:      s.mail,
		Random:    random(42, 43, 44, 45),
		Now:       func() time.Time { return s.now },
	}

	return s
}

// Verifies: DH-1205, DH-1220, DH-1222
func TestAnOTPGoesToTheEmailAddressOfTheCustomerAndIsThenAccepted(t *testing.T) {
	s := newAuthSetup()

	auth, err := s.authenticator.SendOTP(t.Context(), domain.Authentication{}, knownCustomer, "Budget App")

	require.NoError(t, err)
	require.Len(t, s.mail.sent, 1)
	assert.Equal(t, "alex@example.com", s.mail.sent[0].to)
	assert.Equal(t, domain.OTPEmail("000042", "Budget App"), s.mail.sent[0].email)

	auth, result := s.authenticator.Verify(auth, "000042")
	assert.Equal(t, domain.OTPCorrect, result)
	assert.Equal(t, knownCustomer, auth.CustomerNumber())
}

// Verifies: DH-1203
func TestTheEmailAddressInADifferentCaseGetsTheOTPAndACorrectEntryAuthenticatesTheCustomer(t *testing.T) {
	s := newAuthSetup()

	auth, err := s.authenticator.SendOTP(t.Context(), domain.Authentication{}, " Alex@Example.com ", "Budget App")

	require.NoError(t, err)
	require.Len(t, s.mail.sent, 1)
	assert.Equal(t, "alex@example.com", s.mail.sent[0].to)

	auth, result := s.authenticator.Verify(auth, "000042")
	assert.Equal(t, domain.OTPCorrect, result)
	assert.Equal(t, knownCustomer, auth.CustomerNumber())
}

// Verifies: DH-1223
func TestAnEmailAddressThatNoCustomerHasGetsNoEmailAndNoError(t *testing.T) {
	s := newAuthSetup()

	auth, err := s.authenticator.SendOTP(t.Context(), domain.Authentication{}, "sam@example.com", "Budget App")

	require.NoError(t, err)
	assert.Empty(t, s.mail.sent)

	auth, result := s.authenticator.Verify(auth, "000042")
	assert.Equal(t, domain.OTPIncorrect, result)
	assert.Empty(t, auth.CustomerNumber())
}

// Verifies: DH-1225
func TestAnOTPIsRefusedAfter5Minutes(t *testing.T) {
	s := newAuthSetup()
	auth, err := s.authenticator.SendOTP(t.Context(), domain.Authentication{}, knownCustomer, "Budget App")
	require.NoError(t, err)

	s.now = s.now.Add(5 * time.Minute)

	_, result := s.authenticator.Verify(auth, "000042")
	assert.Equal(t, domain.OTPIncorrect, result)
}

// Verifies: DH-1223
func TestACustomerNumberThatDoesNotExistGetsNoEmailAndNoError(t *testing.T) {
	s := newAuthSetup()

	auth, err := s.authenticator.SendOTP(t.Context(), domain.Authentication{}, "99999999", "Budget App")

	require.NoError(t, err)
	assert.Empty(t, s.mail.sent)

	_, result := s.authenticator.Verify(auth, "000042")
	assert.Equal(t, domain.OTPIncorrect, result)
}

// Verifies: DH-1223, DH-1229
func TestACustomerNumberThatDoesNotExistHasTheSameLimitOfOTPs(t *testing.T) {
	s := newAuthSetup()
	auth := domain.Authentication{}

	var err error
	for range 3 {
		auth, err = s.authenticator.SendOTP(t.Context(), auth, "99999999", "Budget App")
		require.NoError(t, err)
	}

	_, err = s.authenticator.SendOTP(t.Context(), auth, "99999999", "Budget App")

	require.ErrorIs(t, err, domain.ErrTooManyOTPs)
}

// Verifies: DH-1229
func TestAFourthOTPIsNotSent(t *testing.T) {
	s := newAuthSetup()
	auth := domain.Authentication{}

	var err error
	for range 3 {
		auth, err = s.authenticator.SendOTP(t.Context(), auth, knownCustomer, "Budget App")
		require.NoError(t, err)
	}

	after, err := s.authenticator.SendOTP(t.Context(), auth, knownCustomer, "Budget App")

	require.ErrorIs(t, err, domain.ErrTooManyOTPs)
	assert.Len(t, s.mail.sent, 3)

	_, result := s.authenticator.Verify(after, "000044")
	assert.Equal(t, domain.OTPCorrect, result, "the third OTP stays valid")
}

func TestAMailFailureGivesNoNewOTPAndAnErrorWithoutTheCause(t *testing.T) {
	s := newAuthSetup()
	s.mail.err = errors.New("550 mailbox alex@example.com is not available")

	auth, err := s.authenticator.SendOTP(t.Context(), domain.Authentication{}, knownCustomer, "Budget App")

	require.ErrorIs(t, err, app.ErrOTPNotSent)
	assert.NotContains(t, err.Error(), "alex@example.com")
	assert.Equal(t, domain.Authentication{}, auth)
}

func TestACustomerStoreFailureIsAnErrorAndSendsNothing(t *testing.T) {
	s := newAuthSetup()
	failure := errors.New("connection lost")
	s.authenticator.Customers = failingCustomers{err: failure}

	_, err := s.authenticator.SendOTP(t.Context(), domain.Authentication{}, knownCustomer, "Budget App")

	require.ErrorIs(t, err, failure)
	assert.Empty(t, s.mail.sent)
}

type failingCustomers struct {
	app.CustomerStore

	err error
}

func (f failingCustomers) Customer(context.Context, string) (app.CustomerRecord, error) {
	return app.CustomerRecord{}, f.err
}

// Verifies: DH-1264, DH-1266
func TestACustomerWithAFixedOTPGetsItByEmailAndIsAuthenticatedWithIt(t *testing.T) {
	s := newAuthSetup()
	fixed, err := domain.ParseFixedOTPs("alex@example.com:111111")
	require.NoError(t, err)

	s.authenticator.FixedOTPs = fixed

	auth, err := s.authenticator.SendOTP(t.Context(), domain.Authentication{}, knownCustomer, "Budget App")

	require.NoError(t, err)
	require.Len(t, s.mail.sent, 1)
	assert.Equal(t, domain.OTPEmail("111111", "Budget App"), s.mail.sent[0].email)

	_, result := s.authenticator.Verify(auth, "111111")
	assert.Equal(t, domain.OTPCorrect, result)
}
