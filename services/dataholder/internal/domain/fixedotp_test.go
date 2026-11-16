package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// Verifies: DH-1264, DH-1265
func TestAFixedOTPReplacesTheRandomOTPOfItsEmailAddressOnly(t *testing.T) {
	fixed, err := domain.ParseFixedOTPs("alex@example.com:111111,sam@example.com:024680")
	require.NoError(t, err)

	assert.Equal(t, "111111", fixed.NewOTP("alex@example.com", 42))
	assert.Equal(t, "024680", fixed.NewOTP("sam@example.com", 42))
	assert.Equal(t, "111111", fixed.NewOTP("Alex@Example.COM", 42), "the case of the address has no effect")
	assert.Equal(t, "000042", fixed.NewOTP("kim@example.com", 42))
}

// Verifies: DH-1267
func TestNoFixedOTPsGivesTheRandomOTP(t *testing.T) {
	fixed, err := domain.ParseFixedOTPs("")
	require.NoError(t, err)

	assert.Equal(t, "000042", fixed.NewOTP("alex@example.com", 42))
	assert.Equal(t, "000042", domain.FixedOTPs{}.NewOTP("alex@example.com", 42))
}

// Verifies: DH-1268
func TestFixedOTPsThatAreNotValidGiveTheEntryNumberAndNoValue(t *testing.T) {
	tests := map[string]struct {
		text string
		want string
	}{
		"no colon":         {text: "alex@example.com:111111,sam@example.com", want: "entry 2"},
		"not an address":   {text: "alex:111111", want: "entry 1"},
		"5 digits":         {text: "alex@example.com:11111", want: "entry 1"},
		"not digits":       {text: "alex@example.com:abcdef", want: "entry 1"},
		"empty entry":      {text: "alex@example.com:111111,", want: "entry 2"},
		"two colons":       {text: "alex@example.com:111111:222222", want: "entry 1"},
		"same address":     {text: "alex@example.com:111111,ALEX@example.com:222222", want: "entry 2"},
		"space in address": {text: " alex@example.com:111111", want: "entry 1"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := domain.ParseFixedOTPs(tt.text)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
			assert.NotContains(t, err.Error(), "alex")
			assert.NotContains(t, err.Error(), "111111")
		})
	}
}
