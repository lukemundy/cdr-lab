package domain_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

var otpStart = time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)

const otpCustomer = "00000001"

// Verifies: DH-1224
func TestAnOTPHas6DecimalDigits(t *testing.T) {
	tests := map[string]struct {
		random uint64
		want   string
	}{
		"zero":                 {random: 0, want: "000000"},
		"leading zeros":        {random: 42, want: "000042"},
		"largest OTP":          {random: 999_999, want: "999999"},
		"first value above":    {random: 1_000_000, want: "000000"},
		"largest random value": {random: math.MaxUint64, want: "551615"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.NewOTP(tt.random))
		})
	}
}

func issued(t *testing.T, otp string, at time.Time) domain.Authentication {
	t.Helper()

	auth, err := domain.Authentication{}.Issue(domain.IssuedOTP{CustomerNumber: otpCustomer, Value: otp}, at)
	require.NoError(t, err)

	return auth
}

// Verifies: DH-1225
func TestTheCorrectOTPIsAcceptedBeforeItsExpiry(t *testing.T) {
	tests := map[string]time.Duration{"at once": 0, "one second before the expiry": 5*time.Minute - time.Second}
	for name, after := range tests {
		t.Run(name, func(t *testing.T) {
			auth := issued(t, "123456", otpStart)

			_, result := auth.Verify("123456", otpStart.Add(after))

			assert.Equal(t, domain.OTPCorrect, result)
		})
	}
}

func TestAnOTPIsRefused(t *testing.T) {
	tests := map[string]struct {
		entry string
		after time.Duration
	}{
		"different digits":      {entry: "654321"},
		"empty":                 {entry: ""},
		"a prefix":              {entry: "12345"},
		"with a space":          {entry: "123456 "},
		"5 minutes after issue": {entry: "123456", after: 5 * time.Minute},
		"a long time after":     {entry: "123456", after: time.Hour},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			auth := issued(t, "123456", otpStart)

			_, result := auth.Verify(tt.entry, otpStart.Add(tt.after))

			assert.Equal(t, domain.OTPIncorrect, result)
		})
	}
}

// Verifies: DH-1222
func TestOnlyACorrectEntryGivesTheCustomerOfTheOTP(t *testing.T) {
	auth := issued(t, "123456", otpStart)
	assert.Empty(t, auth.CustomerNumber(), "before an entry")

	auth, _ = auth.Verify("654321", otpStart)
	assert.Empty(t, auth.CustomerNumber(), "after an incorrect entry")

	auth, result := auth.Verify("123456", otpStart)

	require.Equal(t, domain.OTPCorrect, result)
	assert.Equal(t, otpCustomer, auth.CustomerNumber())
}

func TestAnAuthenticationWithNoCustomerGivesNoCustomerNumber(t *testing.T) {
	auth, err := issued(t, "123456", otpStart).IssueNone(otpStart)
	require.NoError(t, err)

	auth, result := auth.Verify("123456", otpStart)

	assert.Equal(t, domain.OTPIncorrect, result)
	assert.Empty(t, auth.CustomerNumber())
}

// Verifies: DH-1226
func TestAnOTPIsAcceptedOneTimeOnly(t *testing.T) {
	auth := issued(t, "123456", otpStart)

	auth, first := auth.Verify("123456", otpStart)
	_, second := auth.Verify("123456", otpStart)

	assert.Equal(t, domain.OTPCorrect, first)
	assert.Equal(t, domain.OTPIncorrect, second)
}

// Verifies: DH-1228
func TestANewOTPMakesThePreviousOTPNotValid(t *testing.T) {
	auth := issued(t, "111111", otpStart)

	auth, err := auth.Issue(domain.IssuedOTP{CustomerNumber: otpCustomer, Value: "222222"}, otpStart.Add(time.Minute))
	require.NoError(t, err)

	afterOld, old := auth.Verify("111111", otpStart.Add(time.Minute))
	_, current := afterOld.Verify("222222", otpStart.Add(time.Minute))

	assert.Equal(t, domain.OTPIncorrect, old)
	assert.Equal(t, domain.OTPCorrect, current)
}

// Verifies: DH-1225, DH-1228
func TestANewOTPIsValidFor5MinutesFromItsOwnIssue(t *testing.T) {
	auth := issued(t, "111111", otpStart)
	auth, err := auth.Issue(domain.IssuedOTP{CustomerNumber: otpCustomer, Value: "222222"}, otpStart.Add(4*time.Minute))
	require.NoError(t, err)

	_, result := auth.Verify("222222", otpStart.Add(8*time.Minute))

	assert.Equal(t, domain.OTPCorrect, result)
}

// Verifies: DH-1230
func TestTheThirdIncorrectEntryEndsTheAuthentication(t *testing.T) {
	auth := issued(t, "123456", otpStart)

	auth, first := auth.Verify("000001", otpStart)
	auth, second := auth.Verify("000002", otpStart)
	auth, third := auth.Verify("000003", otpStart)
	_, afterTheEnd := auth.Verify("123456", otpStart)

	assert.Equal(t, domain.OTPIncorrect, first)
	assert.Equal(t, domain.OTPIncorrect, second)
	assert.Equal(t, domain.OTPAttemptsUsed, third)
	assert.Equal(t, domain.OTPAttemptsUsed, afterTheEnd, "the correct OTP is not accepted after the end")
}

// Verifies: DH-1230
func TestIncorrectEntriesAreCountedAcrossTheOTPsOfOneFlow(t *testing.T) {
	auth := issued(t, "111111", otpStart)
	auth, _ = auth.Verify("000001", otpStart)
	auth, _ = auth.Verify("000002", otpStart)

	auth, err := auth.Issue(domain.IssuedOTP{CustomerNumber: otpCustomer, Value: "222222"}, otpStart)
	require.NoError(t, err)

	_, result := auth.Verify("000003", otpStart)

	assert.Equal(t, domain.OTPAttemptsUsed, result)
}

// Verifies: DH-1229
func TestAnAuthenticationIssues3OTPsAtMost(t *testing.T) {
	auth := issued(t, "111111", otpStart)

	auth, err := auth.Issue(domain.IssuedOTP{CustomerNumber: otpCustomer, Value: "222222"}, otpStart)
	require.NoError(t, err)
	assert.True(t, auth.CanIssue())

	auth, err = auth.Issue(domain.IssuedOTP{CustomerNumber: otpCustomer, Value: "333333"}, otpStart)
	require.NoError(t, err)
	assert.False(t, auth.CanIssue())

	fourth, err := auth.Issue(domain.IssuedOTP{CustomerNumber: otpCustomer, Value: "444444"}, otpStart)
	require.ErrorIs(t, err, domain.ErrTooManyOTPs)

	_, third := fourth.Verify("333333", otpStart)
	assert.Equal(t, domain.OTPCorrect, third, "the third OTP stays valid after a refused request for a fourth")
}

func TestAnAuthenticationWithNoOTPAcceptsNothing(t *testing.T) {
	_, result := domain.Authentication{}.Verify("", otpStart)

	assert.Equal(t, domain.OTPIncorrect, result)
}

// Verifies: DH-1223
func TestAnAuthenticationOfACustomerThatDoesNotExistAcceptsNothing(t *testing.T) {
	auth, err := domain.Authentication{}.IssueNone(otpStart)
	require.NoError(t, err)

	for _, entry := range []string{"", "000000", "123456"} {
		_, result := auth.Verify(entry, otpStart)
		assert.Equal(t, domain.OTPIncorrect, result, entry)
	}

	assert.True(t, auth.CanIssue())
}

// Verifies: DH-1240, DH-1241, DH-1242
func TestTheOTPEmailGivesTheOTPTheRecipientTheTimeThePurposeAndThePasswordStatement(t *testing.T) {
	email := domain.OTPEmail("042137", "Budget App")

	assert.Equal(t, "Your one-time password to share data with Budget App", email.Subject)
	assert.Equal(t, "Your one-time password is 042137.", email.Heading)
	assert.Equal(t, []string{
		"Use it to share your data with Budget App. It is valid for 5 minutes.",
		"This one-time password is only for the sharing of your data under the Consumer Data Right. " +
			"You cannot use it for a different purpose. Do not give it to a different person.",
		"Budget App does not need your password to get your data. " +
			"Do not give your password to Budget App or to a different person.",
	}, email.Paragraphs)
}

// A device can show the subject on a locked screen.
// Verifies: DH-1243
func TestTheSubjectOfTheOTPEmailDoesNotHaveTheOTP(t *testing.T) {
	assert.NotContains(t, domain.OTPEmail("042137", "Budget App").Subject, "042137")
}

func TestTheDemonstrationNoticeNamesTheBank(t *testing.T) {
	assert.Equal(t,
		"Wattle Bank is not a real bank and this is just a demonstration. "+
			"CDR Lab has no affiliation with any real bank nor the ACCC, DSB, or any real CDR participant.",
		domain.DemonstrationNotice("Wattle Bank"),
	)
}
