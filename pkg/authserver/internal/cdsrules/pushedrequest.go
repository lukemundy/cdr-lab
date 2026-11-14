// Package cdsrules adapts github.com/luikyv/go-oidc to the CDS rules that it does not
// support; it is internal so that services cannot depend on go-oidc types.
package cdsrules

import (
	"context"
	"math"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/luikyv/go-oidc/pkg/goidc"
)

const (
	SharingDurationClaim = "sharing_duration"
	arrangementIDClaim   = "cdr_arrangement_id"
	// RegistrationScope is for registration management only: the CDS forbids it in a token with a data scope.
	RegistrationScope = "cdr:registration"
	// SingleFactorACR is the level of assurance that a one-time password gives.
	SingleFactorACR = "urn:cds.au:cdr:2"
)

const (
	// The CDS uses one year when sharing_duration is longer.
	maxSharingDuration = 365 * 24 * time.Hour
	// The CDS limit between the nbf and the exp of a request object.
	maxRequestObjectLifetime = 60 * time.Minute
)

var requestObjectAlgs = []jose.SignatureAlgorithm{jose.PS256, jose.ES256}

type requestObjectKey struct{}

// requestObject has the members of a request object that go-oidc drops or does not give to a session handler.
type requestObject struct {
	NotBefore *jwt.NumericDate `json:"nbf"`
	Expiry    *jwt.NumericDate `json:"exp"`
	// Claims has the CDS members next to id_token and userinfo.
	Claims map[string]any `json:"claims"`
}

// RequestObjectClaims gives PushedRequest the request object of a PAR request; it does not verify the request object,
// because go-oidc verifies it before it calls a session handler.
func RequestObjectClaims(parPath string) goidc.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != parPath || r.ParseForm() != nil {
				next.ServeHTTP(w, r)
				return
			}

			token, err := jwt.ParseSigned(r.PostForm.Get("request"), requestObjectAlgs)

			var object requestObject
			if err != nil || token.UnsafeClaimsWithoutVerification(&object) != nil {
				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestObjectKey{}, object)))
		})
	}
}

func refused(description string) error {
	return goidc.NewError(goidc.ErrorCodeInvalidRequestObject, description)
}

// PushedRequest is a PAR session handler that applies the CDS rules of a request object; it needs RequestObjectClaims.
func PushedRequest(ctx context.Context, as *goidc.AuthnSession, _ *goidc.Client) error {
	object, _ := ctx.Value(requestObjectKey{}).(requestObject)

	if object.NotBefore != nil && object.Expiry != nil &&
		object.Expiry.Time().Sub(object.NotBefore.Time()) > maxRequestObjectLifetime {
		return refused("exp must not be more than 60 minutes after nbf")
	}

	if slices.Contains(strings.Fields(as.Scopes), RegistrationScope) {
		return goidc.NewError(
			goidc.ErrorCodeInvalidScope,
			"the scope "+RegistrationScope+" is not permitted in an authorisation request",
		)
	}

	if _, ok := object.Claims[arrangementIDClaim]; ok {
		return refused(
			"claims." + arrangementIDClaim + " is not supported: the amendment of a consent is not available",
		)
	}

	if as.Claims != nil && !reachable(as.Claims.IDToken["acr"]) {
		return refused("claims.id_token.acr is essential and does not permit " + SingleFactorACR)
	}

	return sharingDuration(as, object.Claims)
}

// reachable is false only when the acr is essential and names values that do not include SingleFactorACR.
func reachable(acr goidc.ClaimObjectInfo) bool {
	values := acr.Values
	if acr.Value != "" {
		values = append(slices.Clone(values), acr.Value)
	}

	return !acr.IsEssential || len(values) == 0 || slices.Contains(values, SingleFactorACR)
}

// sharingDuration keeps the value in the session store, which go-oidc copies to the grant.
func sharingDuration(as *goidc.AuthnSession, claims map[string]any) error {
	value, ok := claims[SharingDurationClaim]
	if !ok {
		return nil
	}

	seconds, ok := value.(float64)
	if !ok || seconds < 0 || seconds != math.Trunc(seconds) {
		return refused("claims." + SharingDurationClaim + " must be zero or a positive integer")
	}

	if as.Store == nil {
		as.Store = map[string]any{}
	}

	as.Store[SharingDurationClaim] = int(min(seconds, maxSharingDuration.Seconds()))

	return nil
}

// SharingDuration reads the store of a session or a grant, and gives zero when the request has no sharing_duration.
func SharingDuration(store map[string]any) time.Duration {
	return time.Duration(storedSeconds(store[SharingDurationClaim])) * time.Second
}

// A grant store can give a float64 after a round trip through JSON.
func storedSeconds(v any) int {
	switch v := v.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}
