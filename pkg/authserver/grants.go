package authserver

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"

	"github.com/lukemundy/cdr-lab/pkg/authserver/internal/cdsrules"
)

var ErrGrantNotFound = errors.New("authserver: the grant does not exist")

// Grant is the record of the authorisation code and the tokens of one consent; Code and RefreshToken are secrets.
type Grant struct {
	ID       string
	ClientID string
	// Subject is the sub of the consumer for the client.
	Subject       string
	Scopes        []string
	ArrangementID string
	// SharingDuration is zero when the consent is for one data collection with no refresh token.
	SharingDuration time.Duration
	AuthenticatedAt time.Time
	// SharingEndsAt is also the expiry of the refresh token.
	SharingEndsAt time.Time
	CreatedAt     time.Time
	// RevokedAt is zero for a grant that is not revoked.
	RevokedAt time.Time

	Code          string
	CodeExpiresAt time.Time
	// CodeConsumedAt is zero until the client exchanges the code.
	CodeConsumedAt time.Time
	// RedirectURI, Nonce and CodeChallenge are those of the authorisation request.
	RedirectURI   string
	Nonce         string
	CodeChallenge string

	// RefreshToken is empty when the grant has none.
	RefreshToken string
	// CertificateThumbprint is the SHA-256 thumbprint of the client certificate of the last token request.
	CertificateThumbprint string
}

type GrantStore interface {
	// SaveGrant creates the grant, or replaces the grant that has the same ID.
	SaveGrant(ctx context.Context, grant Grant) error
	// Grant, GrantByCode and GrantByRefreshToken give ErrGrantNotFound when the grant does not exist.
	Grant(ctx context.Context, id string) (Grant, error)
	GrantByCode(ctx context.Context, code string) (Grant, error)
	GrantByRefreshToken(ctx context.Context, refreshToken string) (Grant, error)
}

// libraryGrants gives the library the grants of a GrantStore.
type libraryGrants struct{ store GrantStore }

var (
	_ goidc.GrantManager        = libraryGrants{}
	_ goidc.RefreshTokenManager = libraryGrants{}
)

// SaveGrant drops a client credentials grant, which has no code and which no endpoint reads.
func (l libraryGrants) SaveGrant(ctx context.Context, grant *goidc.Grant) error {
	if grant.AuthCode == "" {
		return nil
	}

	consent, sharingDuration := cdsrules.ConsentOf(grant)

	return l.store.SaveGrant(ctx, Grant{
		ID:                    grant.ID,
		ClientID:              grant.ClientID,
		Subject:               consent.Subject,
		Scopes:                strings.Fields(grant.Scopes),
		ArrangementID:         consent.ArrangementID,
		SharingDuration:       sharingDuration,
		AuthenticatedAt:       consent.AuthenticatedAt,
		SharingEndsAt:         consent.SharingEndsAt,
		CreatedAt:             fromUnix(grant.CreatedAt),
		RevokedAt:             fromUnix(grant.RevokedAt),
		Code:                  grant.AuthCode,
		CodeExpiresAt:         fromUnix(grant.AuthCodeExpiresAt),
		CodeConsumedAt:        fromUnix(grant.AuthCodeConsumedAt),
		RedirectURI:           grant.AuthParams.RedirectURI,
		Nonce:                 grant.AuthParams.Nonce,
		CodeChallenge:         grant.AuthParams.CodeChallenge,
		RefreshToken:          grant.RefreshToken,
		CertificateThumbprint: grant.CertThumbprint,
	})
}

func (l libraryGrants) Grant(ctx context.Context, id string) (*goidc.Grant, error) {
	return libraryGrant(l.store.Grant(ctx, id))
}

func (l libraryGrants) GrantByAuthCode(ctx context.Context, code string) (*goidc.Grant, error) {
	return libraryGrant(l.store.GrantByCode(ctx, code))
}

func (l libraryGrants) GrantByRefreshToken(ctx context.Context, refreshToken string) (*goidc.Grant, error) {
	return libraryGrant(l.store.GrantByRefreshToken(ctx, refreshToken))
}

func libraryGrant(grant Grant, err error) (*goidc.Grant, error) {
	if errors.Is(err, ErrGrantNotFound) {
		return nil, goidc.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	out := &goidc.Grant{
		ID:                 grant.ID,
		ClientID:           grant.ClientID,
		Subject:            grant.Subject,
		Scopes:             strings.Join(grant.Scopes, " "),
		CreatedAt:          toUnix(grant.CreatedAt),
		RevokedAt:          toUnix(grant.RevokedAt),
		AuthCode:           grant.Code,
		AuthCodeExpiresAt:  toUnix(grant.CodeExpiresAt),
		AuthCodeConsumedAt: toUnix(grant.CodeConsumedAt),
		AuthParams: goidc.AuthorizationParameters{
			RedirectURI:   grant.RedirectURI,
			Nonce:         grant.Nonce,
			CodeChallenge: grant.CodeChallenge,
			// The data holder accepts no other method.
			CodeChallengeMethod: goidc.CodeChallengeMethodSHA256,
		},
		RefreshToken:   grant.RefreshToken,
		CertThumbprint: grant.CertificateThumbprint,
		Store: cdsrules.StoreOf(cdsrules.Consent{
			ArrangementID:   grant.ArrangementID,
			Subject:         grant.Subject,
			AuthenticatedAt: grant.AuthenticatedAt,
			SharingEndsAt:   grant.SharingEndsAt,
		}, grant.SharingDuration),
	}
	if grant.RefreshToken != "" {
		out.RefreshTokenExpiresAt = toUnix(grant.SharingEndsAt)
	}

	return out, nil
}

// The library uses 0 for a time that is not set.
func fromUnix(seconds int) time.Time {
	if seconds == 0 {
		return time.Time{}
	}

	return time.Unix(int64(seconds), 0).UTC()
}

func toUnix(t time.Time) int {
	if t.IsZero() {
		return 0
	}

	return int(t.Unix())
}

// flows gives the library the sessions of the flows in progress, and the grants of the flows that ended.
type flows struct {
	*sessionStore
	libraryGrants

	requireActive func(ctx context.Context, clientID string) error
}

// GrantByAuthCode checks the status: the exchange has no hook, and a later error revokes the grant.
func (f flows) GrantByAuthCode(ctx context.Context, code string) (*goidc.Grant, error) {
	grant, err := f.libraryGrants.GrantByAuthCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if err := f.requireActive(ctx, grant.ClientID); err != nil {
		return nil, err
	}

	return grant, nil
}

var (
	_ goidc.AuthManager = flows{}
	_ goidc.PARManager  = flows{}
)
