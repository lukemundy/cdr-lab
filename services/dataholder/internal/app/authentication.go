package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// ErrOTPNotSent means that the data holder could not give the OTP to its mail server.
var ErrOTPNotSent = errors.New("the data holder could not send the one-time password")

type Mailer interface {
	Send(ctx context.Context, to string, email domain.Email) error
}

// Authenticator authenticates the consumer of an authorisation flow with a one-time password.
type Authenticator struct {
	Customers CustomerStore
	Mail      Mailer
	// Random gives a value that a person cannot predict.
	Random    func() uint64
	FixedOTPs domain.FixedOTPs
	Now       func() time.Time
}

// SendOTP sends a new OTP to the customer of the user identifier (see Customers.Identify), and gives the
// authentication with that OTP. For an identifier that no customer has, it sends nothing and gives an authentication
// that accepts no entry, with no error. It gives domain.ErrTooManyOTPs at the limit of the flow, and ErrOTPNotSent.
func (a Authenticator) SendOTP(
	ctx context.Context,
	auth domain.Authentication,
	identifier, recipientBrand string,
) (domain.Authentication, error) {
	record, err := identify(ctx, a.Customers, identifier)
	if errors.Is(err, ErrCustomerNotFound) {
		return auth.IssueNone(a.Now())
	}

	if err != nil {
		return auth, fmt.Errorf("read the customer: %w", err)
	}

	otp := a.FixedOTPs.NewOTP(record.Customer.EmailAddress, a.Random())

	next, err := auth.Issue(domain.IssuedOTP{CustomerNumber: record.Customer.CustomerNumber, Value: otp}, a.Now())
	if err != nil {
		return auth, err
	}

	if err := a.Mail.Send(ctx, record.Customer.EmailAddress, domain.OTPEmail(otp, recipientBrand)); err != nil {
		// The cause is not in the error, because a mail error can contain the address of the customer.
		return auth, ErrOTPNotSent
	}

	return next, nil
}

func (a Authenticator) Verify(auth domain.Authentication, entry string) (domain.Authentication, domain.OTPResult) {
	return auth.Verify(entry, a.Now())
}
