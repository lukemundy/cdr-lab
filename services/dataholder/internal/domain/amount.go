package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Amount is AUD in cents, and its JSON is an AmountString with 2 decimal digits.
type Amount int64

// The 16 digits before the full stop are the limit of AmountString.
var centsString = regexp.MustCompile(`^-?[0-9]{1,16}\.[0-9]{2}$`)

// errAmount has no value of the request, because the text can go to telemetry.
var errAmount = errors.New("an amount must be a string with 1 to 16 digits, a full stop and 2 digits")

func (a Amount) String() string {
	sign := ""
	cents := int64(a)

	if cents < 0 {
		sign, cents = "-", -cents
	}

	return fmt.Sprintf("%s%d.%02d", sign, cents/100, cents%100)
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func ParseAmount(text string) (Amount, error) {
	if !centsString.MatchString(text) {
		return 0, errAmount
	}

	cents, err := strconv.ParseInt(strings.Replace(text, ".", "", 1), 10, 64)
	if err != nil {
		return 0, errAmount
	}

	return Amount(cents), nil
}

// AmountBounds gives the smallest Amount that is not less than an AmountString and the largest that is not more,
// because an AmountString can have more than 2 decimal digits.
func AmountBounds(text string) (atLeast, atMost Amount, err error) {
	if !amountString.MatchString(text) {
		return 0, 0, errAmount
	}

	whole, fraction, _ := strings.Cut(text, ".")

	cents, err := ParseAmount(whole + "." + fraction[:2])
	if err != nil {
		return 0, 0, err
	}

	if strings.Trim(fraction[2:], "0") == "" {
		return cents, cents, nil
	}

	if strings.HasPrefix(text, "-") {
		return cents, cents - 1, nil
	}

	return cents + 1, cents, nil
}
