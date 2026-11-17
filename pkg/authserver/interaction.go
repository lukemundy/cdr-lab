package authserver

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"

	"github.com/lukemundy/cdr-lab/pkg/authserver/internal/cdsrules"
)

const (
	interactionPolicyID = "consumer"
	interactionStateKey = "consumer_interaction"
	// The limit from the authorisation request to the end of the flow.
	dataHolderFlowLifetimeSeconds = 600
)

// Interaction is one request of a consumer in an authorisation flow.
type Interaction struct {
	ClientID string
	Scopes   []string
	// SharingDuration is zero when the request has no sharing_duration, and one year at most.
	SharingDuration time.Duration
	// Path is the path on the public listener that the pages of the flow send their forms to. It identifies the flow,
	// so it must not go to telemetry.
	Path string
	// State is the value that the last Outcome of the flow kept. It is nil in the first request of a flow.
	State any
}

// Denial is the error that the software product gets when a flow ends with no consent.
type Denial string

const (
	// AccessDenied means that the consumer stopped the flow or could not authenticate.
	AccessDenied Denial = "access_denied"
	// TemporarilyUnavailable means that the data holder cannot continue the flow now.
	TemporarilyUnavailable Denial = "temporarily_unavailable"
)

type Consent struct {
	// ArrangementID is the cdr_arrangement_id of the token responses.
	ArrangementID string
	// Subject is the final sub of the consumer for the software product: the authorisation server does not change it.
	Subject         string
	AuthenticatedAt time.Time
	// SharingEndsAt is the expiry of the refresh token, when the Interaction has a SharingDuration. No access token
	// expires after it. For an Interaction with no SharingDuration, it is AccessTokenLifetime after the consent.
	SharingEndsAt time.Time
}

// Outcome is the answer of ConsumerPages to one Interaction.
type Outcome struct {
	state   any
	denial  Denial
	consent *Consent
}

// Continue keeps state for the next request of the flow. The caller wrote a page to the response.
func Continue(state any) Outcome {
	return Outcome{state: state}
}

// Deny ends the flow. The authorisation server redirects the browser to the software product with the error, so the
// caller must not write a response.
func Deny(denial Denial) Outcome {
	return Outcome{denial: denial}
}

// Approve ends the flow with an authorisation code for the software product, so the caller must not write a response.
func Approve(consent Consent) Outcome {
	return Outcome{consent: &consent}
}

// Continues gives the state that the outcome keeps, and false when the outcome ends the flow.
func (o Outcome) Continues() (state any, ok bool) {
	return o.state, o.denial == "" && o.consent == nil
}

func (o Outcome) Approved() (Consent, bool) {
	if o.consent == nil {
		return Consent{}, false
	}

	return *o.consent, true
}

// ConsumerPages serves the pages of an authorisation flow.
type ConsumerPages interface {
	Interact(w http.ResponseWriter, r *http.Request, in Interaction) Outcome
	// Refused writes the page for a request that the authorisation server refuses with no redirect to the software
	// product, because it has no authorisation request that it can trust.
	Refused(w http.ResponseWriter, r *http.Request, status int)
}

func consumerPolicy(pages ConsumerPages, requireActive func(context.Context, string) error) goidc.AuthnPolicy {
	return goidc.NewPolicy(
		interactionPolicyID,
		// The CDS permits one use of a request_uri. The library refuses a second use only after the flow ends, so the
		// session loses its pushed request identifier at the first use.
		func(_ *http.Request, session *goidc.AuthnSession, _ *goidc.Client) bool {
			session.PushedAuthReqID = ""

			return true
		},
		func(
			w http.ResponseWriter,
			r *http.Request,
			session *goidc.AuthnSession,
			_ *goidc.Client,
		) (goidc.Status, error) {
			if err := requireActive(r.Context(), session.ClientID); err != nil {
				// The library does not refuse a request for a flow that failed, so the session expires now.
				session.ExpiresAt = 0

				return goidc.StatusFailure, err
			}

			outcome := pages.Interact(w, r, Interaction{
				ClientID:        session.ClientID,
				Scopes:          strings.Fields(session.Scopes),
				SharingDuration: cdsrules.SharingDuration(session.Store),
				Path:            dataHolderAuthorisationPath + "/" + session.ID,
				State:           session.Store[interactionStateKey],
			})
			if outcome.denial != "" {
				// The library does not refuse a request for a flow that failed, so the session expires now.
				session.ExpiresAt = 0

				return goidc.StatusFailure, goidc.NewError(
					goidc.ErrorCode(outcome.denial),
					"the consumer did not give consent",
				)
			}

			if session.Store == nil {
				session.Store = map[string]any{}
			}

			if outcome.consent != nil {
				// The library copies the session store to the grant, so the state of the pages is removed.
				delete(session.Store, interactionStateKey)
				cdsrules.Approve(session, cdsrules.Consent(*outcome.consent))
				// The library does not refuse a request for a flow that ended, so the session expires now.
				session.ExpiresAt = 0

				return goidc.StatusSuccess, nil
			}

			session.Store[interactionStateKey] = outcome.state

			return goidc.StatusPending, nil
		},
	)
}

// refusedPage is the error renderer of the library, for an authorisation request that it cannot answer with a redirect.
func refusedPage(pages ConsumerPages) goidc.RenderErrorFunc {
	return func(w http.ResponseWriter, r *http.Request, err error) error {
		pages.Refused(w, r, refusalStatus(err))

		return nil
	}
}

// The library gives access_denied with no redirect only for a request_uri of a different client, which DSB guidance
// answers with 401.
var unauthorisedRefusals = []goidc.ErrorCode{
	goidc.ErrorCodeAccessDenied,
	goidc.ErrorCodeInvalidClient,
	goidc.ErrorCodeUnauthorizedClient,
}

func refusalStatus(err error) int {
	var refusal goidc.Error
	if !errors.As(err, &refusal) || refusal.Code == goidc.ErrorCodeInternalError {
		return http.StatusInternalServerError
	}

	if slices.Contains(unauthorisedRefusals, refusal.Code) {
		return http.StatusUnauthorized
	}

	return http.StatusBadRequest
}
