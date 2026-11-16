//go:build conformance

package dataholder

import (
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/conformance/internal/labtest"
)

const (
	// noticeProduct has the address of the JWKS server of the tests as its recipient base URI.
	noticeProduct = "5c7e9a1b-3d5f-4b7d-9e1f-2a3b4c5d6e7f"
	// The data holder sends the first attempt at once. The test does not wait for a second attempt.
	waitForNotice = labtest.PollLimit
	pollInterval  = 50 * time.Millisecond
	jwtLifetime   = 300
)

// withdrawn gives the customer and the arrangement after a withdrawal of a consent of the notice product, whose
// notices get the response status.
func withdrawn(t *testing.T, client registeredClient, status int) (account, string) {
	t.Helper()

	customer := newEligibleAccount(t)
	arrangementID := arrangementIDOf(t, customer.consents(client, noAccountScope, ninetyDays).tokens(t))
	labtest.RespondToNotices(arrangementID, status)

	reply := withdrawConsent(t, customer.customer.CustomerNumber, arrangementID)
	require.Equal(t, http.StatusOK, reply.status, string(reply.body))

	return customer, arrangementID
}

func requireNotices(t *testing.T, arrangementID string) []labtest.Notice {
	t.Helper()

	require.Eventually(
		t,
		func() bool { return len(labtest.NoticesFor(arrangementID)) > 0 },
		waitForNotice,
		pollInterval,
		"the software product gets a notice",
	)

	return labtest.NoticesFor(arrangementID)
}

func noticeStatusOf(t *testing.T, customer account, arrangementID string) string {
	t.Helper()

	for _, a := range arrangementsOf(t, customer.customer.CustomerNumber) {
		if a.ArrangementID == arrangementID {
			return a.NoticeStatus
		}
	}

	t.Fatalf("the customer has no arrangement %s", arrangementID)

	return ""
}

func requireNoticeStatus(t *testing.T, customer account, arrangementID, want string) {
	t.Helper()

	require.Eventually(t, func() bool { return noticeStatusOf(t, customer, arrangementID) == want },
		waitForNotice, pollInterval, "the notice becomes %s", want)
}

// Verifies: DH-0903, DH-0924, DH-0925
func TestTheSoftwareProductGetsASignedNoticeAfterAWithdrawal(t *testing.T) {
	client := registered(t, noticeProduct)
	customer, arrangementID := withdrawn(t, client, http.StatusNoContent)
	endpoint := client.reply.body["recipient_base_uri"].(string) + "/arrangements/revoke"
	brandID := labtest.Env(t, "DATAHOLDER_BRAND_ID")

	notices := requireNotices(t, arrangementID)
	notice := notices[0]

	assert.Equal(t, "application/x-www-form-urlencoded", notice.ContentType)
	assert.Equal(t, []string{"cdr_arrangement_jwt"}, slices.Sorted(maps.Keys(notice.Form)), "the body has only the JWT")

	bearer, found := strings.CutPrefix(notice.Authorization, "Bearer ")
	require.True(t, found, notice.Authorization)

	authClaims := client.signedClaims(t, bearer)
	arrangementClaims := client.signedClaims(t, notice.Form.Get("cdr_arrangement_jwt"))

	jwts := map[string]map[string]any{"bearer JWT": authClaims, "cdr_arrangement_jwt": arrangementClaims}
	for name, claims := range jwts {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, brandID, claims["iss"])
			assert.Equal(t, brandID, claims["sub"])
			assert.Equal(t, endpoint, claims["aud"])
			require.IsType(t, "", claims["jti"])
			require.NoError(t, uuid.Validate(claims["jti"].(string)))
			require.IsType(t, float64(0), claims["iat"])
			assert.InDelta(t, claims["iat"].(float64)+jwtLifetime, claims["exp"], 0)
		})
	}

	assert.NotEqual(t, authClaims["jti"], arrangementClaims["jti"])
	assert.Equal(t, arrangementID, arrangementClaims["cdr_arrangement_id"])

	requireNoticeStatus(t, customer, arrangementID, "COMPLETE")
}

func TestTheResponseToANoticeGivesItsStatus(t *testing.T) {
	client := registered(t, noticeProduct)

	tests := map[string]struct {
		response int
		want     string
	}{
		"a 422 response completes the notice":  {response: http.StatusUnprocessableEntity, want: "COMPLETE"},
		"a 400 response makes the notice fail": {response: http.StatusBadRequest, want: "FAILED"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			customer, arrangementID := withdrawn(t, client, tc.response)

			requireNoticeStatus(t, customer, arrangementID, tc.want)
			assert.Len(t, labtest.NoticesFor(arrangementID), 1, "no new attempt")
		})
	}
}

func TestANoticeWithA5xxResponseStaysPendingForANewAttempt(t *testing.T) {
	client := registered(t, noticeProduct)
	customer, arrangementID := withdrawn(t, client, http.StatusServiceUnavailable)

	requireNotices(t, arrangementID)

	assert.Equal(t, "PENDING", noticeStatusOf(t, customer, arrangementID))
}

// Verifies: DH-0902
func TestARevocationByTheSoftwareProductSendsNoNotice(t *testing.T) {
	client := registered(t, noticeProduct)
	customer := newEligibleAccount(t)
	arrangementID := arrangementIDOf(t, customer.consents(client, noAccountScope, ninetyDays).tokens(t))

	r := client.revokeArrangementID(t, arrangementID)
	require.Equal(t, http.StatusNoContent, r.Status, r.Body)

	assert.Empty(t, noticeStatusOf(t, customer, arrangementID), "the arrangement has no notice")
}
