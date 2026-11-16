package problem_test

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
)

// Verifies: PROB-0001
func TestAProblemHasTheStatusPhraseAsItsTitle(t *testing.T) {
	rec := httptest.NewRecorder()

	problem.Write(rec, http.StatusConflict, "The change is not permitted.")

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{"type":"about:blank","title":"Conflict","status":409,"detail":"The change is not permitted."}`,
		rec.Body.String(),
	)
}

func TestAnUnexpectedErrorLogsOnlyTheTypeOfTheError(t *testing.T) {
	var logs bytes.Buffer

	previous := slog.Default()

	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)

	problem.WriteUnexpected(rec, req, errors.New("connection to 10.0.0.1 failed"), "The service failed.")

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(
		t,
		`{"type":"about:blank","title":"Internal Server Error","status":500,"detail":"The service failed."}`,
		rec.Body.String(),
	)
	assert.Contains(t, logs.String(), `"error.type":"*errors.errorString"`)
	assert.NotContains(t, logs.String(), "10.0.0.1")
}

// Verifies: PROB-0002, PROB-0003, PROB-0004, PROB-0006
func TestFieldErrorsHaveTheFieldErrorsType(t *testing.T) {
	rec := httptest.NewRecorder()

	problem.WriteFieldErrors(rec, http.StatusBadRequest, "",
		[]problem.FieldError{{Pointer: problem.FieldPointer("endpoints.website"), Detail: "must be an https URI"}})

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		"type": "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0002",
		"title": "Values in the request are not valid",
		"status": 400,
		"detail": "One or more values in the request are not valid.",
		"errors": [{"pointer": "#/endpoints/website", "detail": "must be an https URI"}]
	}`, rec.Body.String())
}

func TestFieldErrorsKeepTheDetailThatIsGiven(t *testing.T) {
	rec := httptest.NewRecorder()

	problem.WriteFieldErrors(rec, http.StatusConflict,
		"The brand is INACTIVE.", []problem.FieldError{{Pointer: "#/status", Detail: "must be INACTIVE"}})

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), `"detail":"The brand is INACTIVE."`)
}

// Verifies: PROB-0007
func TestNoFieldErrorsIsAProblemWithNoType(t *testing.T) {
	rec := httptest.NewRecorder()

	problem.WriteFieldErrors(rec, http.StatusConflict,
		"The brand is INACTIVE.", nil)

	assert.JSONEq(
		t,
		`{"type":"about:blank","title":"Conflict","status":409,"detail":"The brand is INACTIVE."}`,
		rec.Body.String(),
	)
}

// Verifies: PROB-0010
func TestASessionThatEndedHasTheSessionEndedType(t *testing.T) {
	rec := httptest.NewRecorder()

	problem.WriteSessionEnded(rec, http.StatusForbidden, "Sign in again.")

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		"type": "https://github.com/lukemundy/cdr-lab/blob/main/pkg/platform/problem/DESIGN.md#prob-0010",
		"title": "The session ended",
		"status": 403,
		"detail": "Sign in again."
	}`, rec.Body.String())
}
