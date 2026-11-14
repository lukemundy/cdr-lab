package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func TestAnAmountIsAStringWithTwoDecimalDigits(t *testing.T) {
	tests := map[string]domain.Amount{
		"0.00":                0,
		"1000.00":             100_000,
		"-12.34":              -1234,
		"0.05":                5,
		"9999999999999999.99": 999_999_999_999_999_999,
	}
	for text, want := range tests {
		got, err := domain.ParseAmount(text)
		require.NoError(t, err, text)
		assert.Equal(t, want, got, text)

		data, err := json.Marshal(got)
		require.NoError(t, err)
		assert.JSONEq(t, `"`+text+`"`, string(data))
	}
}

func TestAnAmountThatIsNotAnAmountStringWithTwoDecimalDigitsIsRefused(t *testing.T) {
	for _, text := range []string{
		"1000", "1000.0", "1000.001", "1,000.00", "+1.00", ".50", "1.00 ", "", "10000000000000000.00",
	} {
		_, err := domain.ParseAmount(text)

		require.Error(t, err, text)

		if text != "" {
			assert.NotContains(t, err.Error(), text, "the error has no value of the request")
		}
	}
}

func TestTheBoundsOfAnAmountStringAreTheNearestAmounts(t *testing.T) {
	tests := map[string]struct{ atLeast, atMost domain.Amount }{
		"10.00":    {1000, 1000},
		"10.005":   {1001, 1000},
		"10.0000":  {1000, 1000},
		"-10.005":  {-1000, -1001},
		"-0.001":   {0, -1},
		"0.00":     {0, 0},
		"-12.3400": {-1234, -1234},
	}
	for text, want := range tests {
		atLeast, atMost, err := domain.AmountBounds(text)

		require.NoError(t, err, text)
		assert.Equal(t, want.atLeast, atLeast, text)
		assert.Equal(t, want.atMost, atMost, text)
	}
}

func TestBoundsOfAValueThatIsNotAnAmountStringAreRefused(t *testing.T) {
	for _, text := range []string{"10", "10.0", "1e3", "", "+1.00", "10000000000000000.00"} {
		_, _, err := domain.AmountBounds(text)

		require.Error(t, err, text)
	}
}
