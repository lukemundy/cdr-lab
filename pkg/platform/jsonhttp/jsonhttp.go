// Package jsonhttp reads and writes the JSON bodies of the HTTP interfaces.
package jsonhttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var ErrMoreThanOneValue = errors.New("more than one JSON value")

// Decode refuses a field that v does not have, and a value after the first value. It returns an error of body, for
// example *http.MaxBytesError, as it is.
func Decode(body io.Reader, v any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return err
	}

	switch _, err := decoder.Token(); {
	case errors.Is(err, io.EOF):
		return nil
	case err != nil:
		return err
	default:
		return ErrMoreThanOneValue
	}
}

func Write(w http.ResponseWriter, status int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		// Only a programming error makes a response type that cannot be encoded.
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}
