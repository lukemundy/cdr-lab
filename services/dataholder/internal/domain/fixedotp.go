package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var otpPattern = regexp.MustCompile(fmt.Sprintf(`^[0-9]{%d}$`, otpDigits))

// FixedOTPs is a mock-only feature: each customer with one of its email addresses gets the fixed OTP of that address
// in place of a random OTP, so that browser automation can authenticate. The keys are in lower case.
type FixedOTPs map[string]string

// ParseFixedOTPs reads entries of the form address:otp, separated by commas. An error gives the entry number and no
// value, because an email address is personal information.
func ParseFixedOTPs(text string) (FixedOTPs, error) {
	fixed := FixedOTPs{}
	if text == "" {
		return fixed, nil
	}

	for i, entry := range strings.Split(text, ",") {
		address, otp, found := strings.Cut(entry, ":")
		if !found || !emailAddress.MatchString(address) || !otpPattern.MatchString(otp) {
			return nil, fmt.Errorf(
				"entry %d must be an email address and an OTP of %d digits, separated by a colon", i+1, otpDigits,
			)
		}

		key := strings.ToLower(address)
		if _, ok := fixed[key]; ok {
			return nil, fmt.Errorf("entry %d has the email address of an earlier entry", i+1)
		}

		fixed[key] = otp
	}

	return fixed, nil
}

func (f FixedOTPs) NewOTP(address string, random uint64) string {
	if otp, ok := f[strings.ToLower(address)]; ok {
		return otp
	}

	return NewOTP(random)
}
