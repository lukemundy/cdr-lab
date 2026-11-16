package cdsrules_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver/internal/cdsrules"
)

const year = 365 * 24 * time.Hour

func refreshTokenExpiry(t *testing.T, server *authServer, tokens map[string]any) time.Time {
	t.Helper()

	refresh, ok := tokens["refresh_token"].(string)
	require.True(t, ok, "the token response has a refresh token: %v", tokens)
	_, grant, err := server.op.Introspect(t.Context(), refresh)
	require.NoError(t, err)

	return time.Unix(int64(grant.RefreshTokenExpiresAt), 0)
}

func TestTheRefreshTokenExpiresAtTheEndOfTheSharingDuration(t *testing.T) {
	server, recipient := newAuthServer(t)

	tokens := recipient.exchange(recipient.authorise(cdsClaims(map[string]any{"sharing_duration": 3600})))

	assert.WithinDuration(t, time.Now().Add(time.Hour), refreshTokenExpiry(t, server, tokens), 10*time.Second)
}

// Verifies: AUTH-0090
func TestASharingDurationOfMoreThanOneYearIsOneYear(t *testing.T) {
	server, recipient := newAuthServer(t)

	tokens := recipient.exchange(
		recipient.authorise(cdsClaims(map[string]any{"sharing_duration": int((2 * year).Seconds())})),
	)

	assert.WithinDuration(t, time.Now().Add(year), refreshTokenExpiry(t, server, tokens), 10*time.Second)
}

func TestNoSharingDurationGivesNoRefreshToken(t *testing.T) {
	for name, extra := range map[string]map[string]any{
		"absent": nil,
		"zero":   cdsClaims(map[string]any{"sharing_duration": 0}),
	} {
		t.Run(name, func(t *testing.T) {
			_, recipient := newAuthServer(t)

			tokens := recipient.exchange(recipient.authorise(extra))

			assert.NotContains(t, tokens, "refresh_token")
			assert.NotEmpty(t, tokens["access_token"])
		})
	}
}

// Verifies: AUTH-0089
func TestASharingDurationThatIsNotValidFailsThePushedAuthorisationRequest(t *testing.T) {
	for name, value := range map[string]any{"negative": -1, "not a number": "one year", "fraction": 1.5} {
		t.Run(name, func(t *testing.T) {
			_, recipient := newAuthServer(t)

			a := recipient.authorise(cdsClaims(map[string]any{"sharing_duration": value}))

			requireRefused(t, a.par, "invalid_request_object")
		})
	}
}

// Verifies: AUTH-0093
func TestASharingDurationOutsideTheClaimsMemberHasNoEffect(t *testing.T) {
	_, recipient := newAuthServer(t)

	tokens := recipient.exchange(recipient.authorise(map[string]any{"sharing_duration": 3600}))

	assert.NotContains(t, tokens, "refresh_token")
}

// Verifies: AUTH-0079
func TestARequestObjectThatIsValidFor60MinutesIsAccepted(t *testing.T) {
	_, recipient := newAuthServer(t)
	now := time.Now()

	a := recipient.authorise(map[string]any{
		"nbf": now.Add(-30 * time.Minute).Unix(),
		"exp": now.Add(30 * time.Minute).Unix(),
	})

	require.Equal(t, http.StatusCreated, a.par.status, a.par.body)
}

// Verifies: AUTH-0079
func TestARequestObjectThatIsValidForMoreThan60MinutesIsRefused(t *testing.T) {
	now := time.Now()

	tests := map[string]struct{ nbf, exp time.Time }{
		"61 minutes":  {nbf: now.Add(-31 * time.Minute), exp: now.Add(30 * time.Minute)},
		"118 minutes": {nbf: now.Add(-59 * time.Minute), exp: now.Add(59 * time.Minute)},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, recipient := newAuthServer(t)

			a := recipient.authorise(map[string]any{"nbf": tt.nbf.Unix(), "exp": tt.exp.Unix()})

			requireRefused(t, a.par, "invalid_request_object")
		})
	}
}

// Verifies: AUTH-0091
func TestARequestObjectWithAnArrangementIDIsRefused(t *testing.T) {
	_, recipient := newAuthServer(t)

	a := recipient.authorise(cdsClaims(map[string]any{"cdr_arrangement_id": "5a1bf696-ee03-408b-b315-97955415d1f0"}))

	requireRefused(t, a.par, "invalid_request_object")
}

// Verifies: AUTH-0086
func TestARequestForTheRegistrationScopeIsRefused(t *testing.T) {
	_, recipient := newAuthServer(t)

	a := recipient.authorise(map[string]any{"scope": "openid " + dataScope + " " + cdsrules.RegistrationScope})

	requireRefused(t, a.par, "invalid_scope")
}

const loa2, loa3 = "urn:cds.au:cdr:2", "urn:cds.au:cdr:3"

func acrClaim(acr map[string]any) map[string]any {
	return cdsClaims(map[string]any{"id_token": map[string]any{"acr": acr}})
}

// Verifies: AUTH-0092
func TestAnEssentialACRThatTheDataHolderCannotReachIsRefused(t *testing.T) {
	tests := map[string]map[string]any{
		"essential value 3":  {"essential": true, "value": loa3},
		"essential values 3": {"essential": true, "values": []string{loa3}},
	}
	for name, acr := range tests {
		t.Run(name, func(t *testing.T) {
			_, recipient := newAuthServer(t)

			a := recipient.authorise(acrClaim(acr))

			requireRefused(t, a.par, "invalid_request_object")
		})
	}
}

// Verifies: AUTH-0092
func TestAnACRThatTheDataHolderReachesOrThatIsVoluntaryIsAccepted(t *testing.T) {
	tests := map[string]map[string]any{
		"essential value 2":        {"essential": true, "value": loa2},
		"essential values 2 and 3": {"essential": true, "values": []string{loa3, loa2}},
		"essential with no values": {"essential": true},
		"voluntary value 3":        {"value": loa3},
	}
	for name, acr := range tests {
		t.Run(name, func(t *testing.T) {
			_, recipient := newAuthServer(t)

			a := recipient.authorise(acrClaim(acr))

			require.Equal(t, http.StatusCreated, a.par.status, a.par.body)
		})
	}
}

// cdsClaims gives the claims member of a request object, where the CDS puts sharing_duration and cdr_arrangement_id.
func cdsClaims(members map[string]any) map[string]any {
	return map[string]any{"claims": members}
}

func requireRefused(t *testing.T, par reply, code string) {
	t.Helper()
	require.Equal(t, http.StatusBadRequest, par.status, par.body)
	assert.Equal(t, code, par.body["error"])
}
