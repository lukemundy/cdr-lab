package domain

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"time"
)

const (
	otpDigits = 6
	// OTPLifetime is the time for which an OTP is valid.
	OTPLifetime = 5 * time.Minute
	// The limits for one authorisation flow.
	maxOTPs             = 3
	maxIncorrectEntries = 3
)

// otpValues is the count of OTPs with otpDigits decimal digits.
const otpValues = 1_000_000

var ErrTooManyOTPs = errors.New("the authentication issued the maximum number of one-time passwords")

// NewOTP makes a one-time password from a random value, which the app layer gives.
func NewOTP(random uint64) string {
	return fmt.Sprintf("%0*d", otpDigits, random%otpValues)
}

type OTPResult int

const (
	OTPIncorrect OTPResult = iota
	OTPCorrect
	// OTPAttemptsUsed means that the authentication had its last incorrect entry, and accepts no entry.
	OTPAttemptsUsed
)

// Authentication is the state of the one-time passwords of one authorisation flow.
type Authentication struct {
	// otp is empty when no entry can be correct: before the first OTP, after a correct entry, and for a customer
	// that does not exist.
	otp string
	// customerNumber is the customer of the OTP, and is empty for a customer that does not exist.
	customerNumber string
	authenticated  bool
	expiresAt      time.Time
	issued         int
	incorrect      int
}

func (a Authentication) CanIssue() bool {
	return a.issued < maxOTPs
}

// IssuedOTP is an OTP for one customer: a correct entry authenticates that customer.
type IssuedOTP struct {
	CustomerNumber string
	Value          string
}

// Issue replaces the OTP of the authentication. It gives ErrTooManyOTPs and no change at the limit.
func (a Authentication) Issue(otp IssuedOTP, now time.Time) (Authentication, error) {
	if !a.CanIssue() {
		return a, ErrTooManyOTPs
	}

	a.customerNumber, a.otp, a.expiresAt = otp.CustomerNumber, otp.Value, now.Add(OTPLifetime)
	a.issued++

	return a, nil
}

// IssueNone counts an OTP and keeps none, for a customer that does not exist: the flow must not show the difference.
func (a Authentication) IssueNone(now time.Time) (Authentication, error) {
	return a.Issue(IssuedOTP{}, now)
}

// CustomerNumber is the customer that a correct entry authenticated, and is empty before a correct entry.
func (a Authentication) CustomerNumber() string {
	if !a.authenticated {
		return ""
	}

	return a.customerNumber
}

func (a Authentication) Verify(entry string, now time.Time) (Authentication, OTPResult) {
	if a.incorrect >= maxIncorrectEntries {
		return a, OTPAttemptsUsed
	}

	valid := a.otp != "" && now.Before(a.expiresAt)
	if valid && subtle.ConstantTimeCompare([]byte(entry), []byte(a.otp)) == 1 {
		a.otp, a.authenticated = "", true

		return a, OTPCorrect
	}

	a.incorrect++
	if a.incorrect >= maxIncorrectEntries {
		return a, OTPAttemptsUsed
	}

	return a, OTPIncorrect
}

// Email is a message to one customer. Each part of the email gives the heading and then each paragraph.
type Email struct {
	Subject    string
	Heading    string
	Paragraphs []string
}

// OTPEmail gives the message that takes an OTP to a customer; recipient is the brand name of the data recipient.
func OTPEmail(otp, recipient string) Email {
	return Email{
		Subject: "Your one-time password to share data with " + recipient,
		Heading: "Your one-time password is " + otp + ".",
		Paragraphs: []string{
			fmt.Sprintf(
				"Use it to share your data with %s. It is valid for %d minutes.", recipient, int(OTPLifetime.Minutes()),
			),
			"This one-time password is only for the sharing of your data under the Consumer Data Right. " +
				"You cannot use it for a different purpose. Do not give it to a different person.",
			recipient + " does not need your password to get your data. " +
				"Do not give your password to " + recipient + " or to a different person.",
		},
	}
}

// DemonstrationNotice is the footer text of each page and email; bank is the brand name of the data holder.
func DemonstrationNotice(bank string) string {
	return bank + " is not a real bank and this is just a demonstration. " +
		"CDR Lab has no affiliation with any real bank nor the ACCC, DSB, or any real CDR participant."
}
