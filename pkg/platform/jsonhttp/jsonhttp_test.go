package jsonhttp_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
)

type request struct {
	Name string `json:"name"`
}

func TestDecodeAcceptsOneValueWithOnlyTheFieldsOfTheRequest(t *testing.T) {
	var got request

	require.NoError(t, jsonhttp.Decode(strings.NewReader(` {"name": "a"} `+"\n"), &got))
	assert.Equal(t, request{Name: "a"}, got)
}

func TestDecodeRefusesABodyThatIsNotOneValueWithOnlyTheFieldsOfTheRequest(t *testing.T) {
	tests := map[string]struct {
		body string
		want func(t *testing.T, err error)
	}{
		"empty": {"", func(t *testing.T, err error) { t.Helper(); assert.ErrorIs(t, err, io.EOF) }},
		"not JSON": {"{", func(t *testing.T, err error) {
			t.Helper()

			var syntaxErr *json.SyntaxError

			assert.True(t, errors.As(err, &syntaxErr) || errors.Is(err, io.ErrUnexpectedEOF), err)
		}},
		// Verifies: BODY-0001
		"an unknown field": {`{"name": "a", "other": 1}`, func(t *testing.T, err error) {
			t.Helper()
			assert.ErrorContains(t, err, `unknown field "other"`)
		}},
		// Verifies: BODY-0002
		"two values": {`{"name": "a"} {"name": "b"}`, func(t *testing.T, err error) {
			t.Helper()
			assert.ErrorIs(t, err, jsonhttp.ErrMoreThanOneValue)
		}},
		// Verifies: BODY-0002
		"a value after the object": {`{"name": "a"} 1`, func(t *testing.T, err error) {
			t.Helper()
			assert.ErrorIs(t, err, jsonhttp.ErrMoreThanOneValue)
		}},
		"a delimiter after the object": {`{"name": "a"}}`, func(t *testing.T, err error) {
			t.Helper()
			assert.Error(t, err)
		}},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var got request

			tt.want(t, jsonhttp.Decode(strings.NewReader(tt.body), &got))
		})
	}
}

func TestDecodeGivesTheErrorOfTheReader(t *testing.T) {
	rec := httptest.NewRecorder()
	body := http.MaxBytesReader(rec, io.NopCloser(strings.NewReader(`{"name": "abcdef"}`)), 4)

	var got request

	var tooLarge *http.MaxBytesError

	require.ErrorAs(t, jsonhttp.Decode(body, &got), &tooLarge)
}

// Verifies: BODY-0004
func TestWriteGivesTheBodyAsJSONWithTheStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	jsonhttp.Write(rec, http.StatusCreated, request{Name: "a"})

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"name": "a"}`, rec.Body.String())
}
