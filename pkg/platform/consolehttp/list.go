package consolehttp

import (
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strconv"
)

// The list parameters of a console API. The error messages do not contain the values, because a value can be the
// data of a participant or a customer.

const (
	defaultPageSize = 25
	maxPageSize     = 100
)

func ParsePage(v url.Values) (page, pageSize int, err error) {
	if page, err = positiveInt(v, "page", 1); err != nil {
		return 0, 0, err
	}

	if pageSize, err = positiveInt(v, "pageSize", defaultPageSize); err != nil {
		return 0, 0, err
	}

	if pageSize > maxPageSize {
		return 0, 0, fmt.Errorf("pageSize must be %d or less", maxPageSize)
	}

	return page, pageSize, nil
}

// ParseEnum gives nil when the parameter is absent, which is no filter.
func ParseEnum[T ~string](v url.Values, name string, allowed []T) (*T, error) {
	s := v.Get(name)
	if s == "" {
		return nil, nil //nolint:nilnil // nil is no filter, which is not an error.
	}

	value := T(s)
	if !slices.Contains(allowed, value) {
		return nil, fmt.Errorf("%s must be one of %v", name, allowed)
	}

	return &value, nil
}

func ParseSort[T any](v url.Values, sorts map[string]T, fallback T) (T, error) {
	s := v.Get("sort")
	if s == "" {
		return fallback, nil
	}

	sort, ok := sorts[s]
	if !ok {
		return fallback, fmt.Errorf("sort must be one of %v", slices.Sorted(maps.Keys(sorts)))
	}

	return sort, nil
}

func ParseOrder(v url.Values) (descending bool, err error) {
	switch v.Get("order") {
	case "", "asc":
		return false, nil
	case "desc":
		return true, nil
	default:
		return false, errors.New("order must be asc or desc")
	}
}

func positiveInt(v url.Values, name string, fallback int) (int, error) {
	s := v.Get(name)
	if s == "" {
		return fallback, nil
	}

	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%s must be an integer of 1 or more", name)
	}

	return n, nil
}
