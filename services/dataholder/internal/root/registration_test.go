package root

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/pkg/platform/jobs"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/grantstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/jtistore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const testIssuer = "https://id.bank.example"

// verifiedJWTs gives a valid SSA and a valid request, each with a new jti for each request JWT.
type verifiedJWTs struct{ now time.Time }

func (verifiedJWTs) SoftwareStatement(requestJWT string) (string, error) {
	return "ssa-of-" + requestJWT, nil
}

func (j verifiedJWTs) VerifySSA(_ context.Context, ssa string) (domain.SSA, error) {
	return domain.SSA{
		Issuer:            "cdr-register",
		IssuedAt:          j.now.Unix(),
		ExpiresAt:         j.now.Add(time.Minute).Unix(),
		ID:                ssa,
		OrgID:             "brand-1",
		OrgName:           "Recipient",
		ClientName:        "Recipient Web",
		ClientDescription: "A software product",
		ClientURI:         "https://recipient.example",
		RedirectURIs:      []string{"https://recipient.example/cb"},
		LogoURI:           "https://recipient.example/logo.png",
		JWKSURI:           "https://recipient.example/jwks",
		RevocationURI:     "https://recipient.example/revoke",
		RecipientBaseURI:  "https://recipient.example",
		SoftwareID:        "software-1",
		SoftwareRoles:     "data-recipient-software-product",
		Scope:             "openid cdr:registration",
	}, nil
}

func (j verifiedJWTs) VerifyRequest(_ context.Context, requestJWT, _ string) (domain.RegistrationRequest, error) {
	//nolint:gosec // G101 reads the protocol field token_endpoint_auth_method as a credential
	return domain.RegistrationRequest{
		Issuer:                         "software-1",
		IssuedAt:                       j.now.Unix(),
		ExpiresAt:                      j.now.Add(time.Minute).Unix(),
		ID:                             requestJWT,
		Audience:                       []string{testIssuer},
		TokenEndpointAuthMethod:        "private_key_jwt",
		TokenEndpointAuthSigningAlg:    "PS256",
		GrantTypes:                     []string{"client_credentials"},
		ResponseTypes:                  []string{"code"},
		IDTokenSignedResponseAlg:       "PS256",
		AuthorizationSignedResponseAlg: "PS256",
		SoftwareStatement:              "ssa-of-" + requestJWT,
	}, nil
}

// startedRegistrar gives a registrar with new stores, as a start of the data holder makes it.
func startedRegistrar(t *testing.T, pool *pgxpool.Pool) app.Registrar {
	t.Helper()

	now := time.Now()
	inserter, err := jobs.NewInsertOnlyClient(pool)
	require.NoError(t, err)

	events := newEventStore(pool, inserter)

	return app.Registrar{
		Issuer:        testIssuer,
		JWTs:          verifiedJWTs{now: now},
		JTIs:          jtistore.New(pool),
		Store:         events,
		Registrations: registrationstore.NewQueries(pool),
		Statuses:      statusstore.New(pool, nil),
		Arrangements: app.Arrangements{
			Store: events, Grants: grantstore.New(pool), Reader: arrangementstore.NewQueries(pool), Now: time.Now,
		},
		Now:   func() time.Time { return now },
		NewID: uuid.NewString,
	}
}

func TestARegistrationStaysAfterARestart(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	first, err := startedRegistrar(t, pool).Register(t.Context(), "request-1")
	require.NoError(t, err)
	assert.NotEmpty(t, first.ClientID)

	_, err = startedRegistrar(t, pool).Register(t.Context(), "request-2")

	var refusal domain.RegistrationError
	require.ErrorAs(t, err, &refusal)
	assert.Equal(t, domain.InvalidSoftwareStatement, refusal.Code)
}

func TestTwoRegistrationsOfOneSoftwareProductAtTheSameTimeGiveOneRegistration(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	registrar := startedRegistrar(t, pool)
	errs := make(chan error, 2)

	for _, request := range []string{"request-1", "request-2"} {
		go func() {
			_, err := registrar.Register(t.Context(), request)
			errs <- err
		}()
	}

	first, second := <-errs, <-errs
	if first == nil {
		first, second = second, first
	}

	require.NoError(t, second, "one registration succeeds")

	var refusal domain.RegistrationError
	require.ErrorAs(t, first, &refusal)
	assert.Equal(t, domain.InvalidSoftwareStatement, refusal.Code)
}

func TestAnUpdateAndADeletionStayAfterARestart(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	first, err := startedRegistrar(t, pool).Register(t.Context(), "request-1")
	require.NoError(t, err)

	updated, err := startedRegistrar(t, pool).Update(t.Context(), first.ClientID, "request-2")
	require.NoError(t, err)

	afterUpdate, err := startedRegistrar(t, pool).Registrations.Registration(t.Context(), first.ClientID)
	require.NoError(t, err)
	assert.Equal(t, updated, afterUpdate)

	require.NoError(t, startedRegistrar(t, pool).Delete(t.Context(), first.ClientID))

	_, err = startedRegistrar(t, pool).Registrations.Registration(t.Context(), first.ClientID)
	require.ErrorIs(t, err, app.ErrRegistrationNotFound)

	again, err := startedRegistrar(t, pool).Register(t.Context(), "request-3")
	require.NoError(t, err)
	assert.NotEqual(t, first.ClientID, again.ClientID)
}
