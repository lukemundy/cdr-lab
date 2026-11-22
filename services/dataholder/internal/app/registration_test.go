package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const (
	issuer     = "https://id.bank.example"
	requestJWT = "header.request.signature"
	ssaJWT     = "header.ssa.signature"
)

var registrationTime = time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

// memoryStore keeps streams in memory, with the version check of a real store.
type memoryStore struct {
	streams map[string][]eventsourcing.RecordedEvent
}

func (s *memoryStore) Read(_ context.Context, streamType, streamID string) ([]eventsourcing.RecordedEvent, error) {
	return s.streams[streamType+"/"+streamID], nil
}

func (s *memoryStore) Append(_ context.Context, appends ...eventsourcing.Append) error {
	if s.streams == nil {
		s.streams = map[string][]eventsourcing.RecordedEvent{}
	}

	for _, a := range appends {
		key := a.StreamType + "/" + a.StreamID
		if int64(len(s.streams[key])) != a.ExpectedVersion {
			return eventsourcing.ErrVersionConflict
		}

		for _, e := range a.Events {
			s.streams[key] = append(s.streams[key], eventsourcing.RecordedEvent{
				EventData: e, StreamVersion: int64(len(s.streams[key]) + 1),
			})
		}
	}

	return nil
}

// jwts gives the claims of the fixed request JWT and SSA, or the error of each step.
type jwts struct {
	ssa        domain.SSA
	request    domain.RegistrationRequest
	readErr    error
	ssaErr     error
	requestErr error
	// keysFrom is the JWKS URI that the request was verified with.
	keysFrom string
}

func (j *jwts) SoftwareStatement(string) (string, error) {
	return j.request.SoftwareStatement, j.readErr
}

func (j *jwts) VerifySSA(context.Context, string) (domain.SSA, error) {
	return j.ssa, j.ssaErr
}

func (j *jwts) VerifyRequest(_ context.Context, _, jwksURI string) (domain.RegistrationRequest, error) {
	j.keysFrom = jwksURI

	return j.request, j.requestErr
}

type sectorDocuments struct {
	redirectURIs []string
	err          error
	requested    []string
}

func (s *sectorDocuments) RedirectURIs(_ context.Context, uri string) ([]string, error) {
	s.requested = append(s.requested, uri)

	return s.redirectURIs, s.err
}

type jtis map[string]bool

func (j jtis) Consume(_ context.Context, jti string) error {
	if j[jti] {
		return app.ErrJTIUsed
	}

	j[jti] = true

	return nil
}

func (j jtis) ConsumeSSA(ctx context.Context, jti string, _ time.Time) error {
	return j.Consume(ctx, "ssa/"+jti)
}

func validJWTs() *jwts {
	return &jwts{
		ssa: domain.SSA{
			Issuer:            "cdr-register",
			IssuedAt:          registrationTime.Add(-time.Minute).Unix(),
			ExpiresAt:         registrationTime.Add(9 * time.Minute).Unix(),
			ID:                "ssa-jti",
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
		},
		//nolint:gosec // G101 reads the protocol field token_endpoint_auth_method as a credential
		request: domain.RegistrationRequest{
			Issuer:                         "software-1",
			IssuedAt:                       registrationTime.Unix(),
			ExpiresAt:                      registrationTime.Add(time.Minute).Unix(),
			ID:                             "request-jti",
			Audience:                       []string{issuer},
			TokenEndpointAuthMethod:        "private_key_jwt",
			TokenEndpointAuthSigningAlg:    "PS256",
			GrantTypes:                     []string{"client_credentials"},
			ResponseTypes:                  []string{"code"},
			IDTokenSignedResponseAlg:       "PS256",
			AuthorizationSignedResponseAlg: "PS256",
			SoftwareStatement:              ssaJWT,
		},
	}
}

func newRegistrar(j *jwts) app.Registrar {
	ids := 0

	store := &memoryStore{}

	return app.Registrar{
		Issuer:        issuer,
		JWTs:          j,
		JTIs:          jtis{},
		Store:         store,
		Registrations: storeReader{store: store},
		Statuses:      softwareStatuses{},
		Arrangements:  &clientArrangements{},
		Sectors:       &sectorDocuments{redirectURIs: []string{"https://recipient.example/cb"}},
		Now:           func() time.Time { return registrationTime },
		ClockLeeway:   30 * time.Second,
		NewID: func() string {
			ids++
			return fmt.Sprintf("client-%d", ids)
		},
	}
}

// clientArrangements records each end of the arrangements of a deleted registration, or fails with err.
type clientArrangements struct {
	ends []string
	err  error
}

func (c *clientArrangements) EndForDeletedRegistration(
	_ context.Context,
	softwareID, clientID string,
	deletedAt time.Time,
) error {
	if c.err != nil {
		return c.err
	}

	c.ends = append(c.ends, softwareID+"/"+clientID+" at "+deletedAt.Format(time.RFC3339))

	return nil
}

// storeReader finds a registration in the streams of a memoryStore, as the read model does for a real store.
type storeReader struct{ store *memoryStore }

func (r storeReader) Registration(_ context.Context, clientID string) (domain.Registration, error) {
	for key, recorded := range r.store.streams {
		state, _, err := app.ClientRegistrationType.Fold(key, recorded)
		if err != nil {
			return domain.Registration{}, err
		}

		if state.Registered && state.Registration.ClientID == clientID {
			return state.Registration, nil
		}
	}

	return domain.Registration{}, app.ErrRegistrationNotFound
}

// conflictingStore gives a version conflict for its first appends, as a store does when a different request changed
// the stream.
type conflictingStore struct {
	*memoryStore

	conflicts int
}

func (s *conflictingStore) Append(ctx context.Context, appends ...eventsourcing.Append) error {
	if s.conflicts > 0 {
		s.conflicts--
		return eventsourcing.ErrVersionConflict
	}

	return s.memoryStore.Append(ctx, appends...)
}

func requireRegistrationError(t *testing.T, err error, code domain.RegistrationErrorCode, description string) {
	t.Helper()

	var registrationErr domain.RegistrationError
	require.ErrorAs(t, err, &registrationErr)
	assert.Equal(t, code, registrationErr.Code)
	assert.Contains(t, registrationErr.Description, description)
}

// Verifies: DH-0633
func TestAValidRequestRegistersTheSoftwareProductWithANewClientID(t *testing.T) {
	j := validJWTs()

	registration, err := newRegistrar(j).Register(t.Context(), requestJWT)

	require.NoError(t, err)
	assert.Equal(t, "client-1", registration.ClientID)
	assert.Equal(t, registrationTime.Unix(), registration.ClientIDIssuedAt)
	assert.Equal(t, "software-1", registration.SoftwareID)
	assert.Equal(t, ssaJWT, registration.SoftwareStatement)
	assert.Equal(t, "https://recipient.example/jwks", j.keysFrom, "the keys come from the jwks_uri of the SSA")
}

// Verifies: DH-0632, DH-0651
func TestASecondRegistrationOfTheSameSoftwareProductIsRefused(t *testing.T) {
	j := validJWTs()
	registrar := newRegistrar(j)
	_, err := registrar.Register(t.Context(), requestJWT)
	require.NoError(t, err)

	j.request.ID, j.ssa.ID = "request-jti-2", "ssa-jti-2"
	_, err = registrar.Register(t.Context(), requestJWT)

	requireRegistrationError(t, err, domain.InvalidSoftwareStatement, "software_id")
}

// Verifies: DH-0620
func TestAnSSAThatTheDataHolderAcceptedBeforeIsRefused(t *testing.T) {
	j := validJWTs()
	registrar := newRegistrar(j)
	_, err := registrar.Register(t.Context(), requestJWT)
	require.NoError(t, err)

	registrar.Store = &memoryStore{}
	j.request.ID = "request-jti-2"
	_, err = registrar.Register(t.Context(), requestJWT)

	requireRegistrationError(t, err, domain.InvalidSoftwareStatement, "jti of the SSA")
}

// Verifies: DH-0620
func TestTheSSAJTIIsKeptUntilTheEndOfTheClockLeewayAfterTheSSAExpires(t *testing.T) {
	j := validJWTs()
	store := &expiryJTIs{}
	registrar := newRegistrar(j)
	registrar.JTIs = store

	_, err := registrar.Register(t.Context(), requestJWT)

	require.NoError(t, err)
	assert.Equal(t, time.Unix(j.ssa.ExpiresAt, 0).Add(30*time.Second), store.ssaExpiresAt)
}

// Verifies: DH-0620
func TestAnSSAThatTheDataHolderAcceptedBeforeIsRefusedInTheClockLeewayAfterItExpires(t *testing.T) {
	j := validJWTs()
	at := registrationTime
	registrar := newRegistrar(j)
	registrar.Now = func() time.Time { return at }
	registrar.JTIs = &expiringJTIs{now: registrar.Now, ssa: map[string]time.Time{}}
	_, err := registrar.Register(t.Context(), requestJWT)
	require.NoError(t, err)

	at = time.Unix(j.ssa.ExpiresAt, 0).Add(30 * time.Second)
	registrar.Store = &memoryStore{}
	j.request.ID = "request-jti-2"
	j.request.ExpiresAt = at.Add(time.Minute).Unix()
	_, err = registrar.Register(t.Context(), requestJWT)

	requireRegistrationError(t, err, domain.InvalidSoftwareStatement, "jti of the SSA")
}

// expiringJTIs keeps each SSA JWT ID until its expiry time, as the JWT ID store of the data holder does.
type expiringJTIs struct {
	now func() time.Time
	ssa map[string]time.Time
}

func (*expiringJTIs) Consume(context.Context, string) error { return nil }

func (e *expiringJTIs) ConsumeSSA(_ context.Context, jti string, expiresAt time.Time) error {
	if kept, ok := e.ssa[jti]; ok && !e.now().After(kept) {
		return app.ErrJTIUsed
	}

	e.ssa[jti] = expiresAt

	return nil
}

// expiryJTIs keeps the expiry time of the last SSA JWT ID.
type expiryJTIs struct{ ssaExpiresAt time.Time }

func (*expiryJTIs) Consume(context.Context, string) error { return nil }

func (e *expiryJTIs) ConsumeSSA(_ context.Context, _ string, expiresAt time.Time) error {
	e.ssaExpiresAt = expiresAt

	return nil
}

// Verifies: DH-0693, DH-0655
func TestARegistrationOfARemovedSoftwareProductIsRefused(t *testing.T) {
	j := validJWTs()
	registrar := newRegistrar(j)
	registrar.Statuses = softwareStatuses{"software-1": domain.SoftwareProductRemoved}

	_, err := registrar.Register(t.Context(), requestJWT)

	requireRegistrationError(t, err, domain.UnapprovedSoftwareStatement, "REMOVED")
	assert.Empty(t, j.keysFrom, "the data holder does not get the keys of a REMOVED software product")
}

// Verifies: DH-0621
func TestASoftwareProductThatIsNotRemovedCanRegister(t *testing.T) {
	tests := map[string]softwareStatuses{
		"ACTIVE":    {"software-1": domain.SoftwareProductActive},
		"INACTIVE":  {"software-1": domain.SoftwareProductInactive},
		"no status": {},
	}
	for name, statuses := range tests {
		t.Run(name, func(t *testing.T) {
			registrar := newRegistrar(validJWTs())
			registrar.Statuses = statuses

			_, err := registrar.Register(t.Context(), requestJWT)

			assert.NoError(t, err)
		})
	}
}

func TestAFailureToReadTheStatusIsNotARegistrationError(t *testing.T) {
	j := validJWTs()
	j.ssa.SoftwareID, j.request.Issuer = "software-failure", "software-failure"

	_, err := newRegistrar(j).Register(t.Context(), requestJWT)

	require.Error(t, err)

	var registrationErr domain.RegistrationError
	assert.NotErrorAs(t, err, &registrationErr)
}

// Verifies: DH-0608
func TestARequestJWTThatWasUsedBeforeIsRefused(t *testing.T) {
	j := validJWTs()
	registrar := newRegistrar(j)
	_, err := registrar.Register(t.Context(), requestJWT)
	require.NoError(t, err)

	registrar.Store = &memoryStore{}
	_, err = registrar.Register(t.Context(), requestJWT)

	requireRegistrationError(t, err, domain.InvalidClientMetadata, "jti")
}

func TestARequestThatIsNotValidIsRefusedWithTheCodeOfItsFault(t *testing.T) {
	tests := map[string]struct {
		change      func(*jwts)
		code        domain.RegistrationErrorCode
		description string
	}{
		"a body that is not a JWT": {
			func(j *jwts) { j.readErr = app.ErrInvalidJWT }, domain.InvalidClientMetadata, "not a JWT",
		},
		"no software statement": {
			func(j *jwts) { j.request.SoftwareStatement = "" }, domain.InvalidSoftwareStatement, "software_statement",
		},
		"an SSA that does not verify": {
			func(j *jwts) { j.ssaErr = app.ErrInvalidJWT }, domain.InvalidSoftwareStatement, "signature",
		},
		"an SSA with a claim that is not valid": {
			func(j *jwts) { j.ssa.Issuer = "other" }, domain.InvalidSoftwareStatement, "iss",
		},
		"a request that does not verify": {
			func(j *jwts) { j.requestErr = app.ErrInvalidJWT }, domain.InvalidClientMetadata, "signature",
		},
		"keys of the software product that are not available": {
			func(j *jwts) { j.requestErr = app.ErrKeysNotAvailable }, domain.InvalidClientMetadata, "jwks_uri",
		},
		"a request with a claim that is not valid": {
			func(j *jwts) { j.request.Audience = nil }, domain.InvalidClientMetadata, "aud",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			j := validJWTs()
			tc.change(j)

			_, err := newRegistrar(j).Register(t.Context(), requestJWT)

			requireRegistrationError(t, err, tc.code, tc.description)
		})
	}
}

func TestAFailureToVerifyTheSSAThatIsNotTheFaultOfTheRequestIsNotARegistrationError(t *testing.T) {
	j := validJWTs()
	j.ssaErr = errors.New("the Register JWKS is not available")

	_, err := newRegistrar(j).Register(t.Context(), requestJWT)

	require.Error(t, err)

	var registrationErr domain.RegistrationError
	assert.NotErrorAs(t, err, &registrationErr)
}

func TestARequestThatIsRefusedDoesNotUseItsJTI(t *testing.T) {
	j := validJWTs()
	registrar := newRegistrar(j)
	j.request.TokenEndpointAuthMethod = "client_secret_basic"
	_, err := registrar.Register(t.Context(), requestJWT)
	require.Error(t, err)

	j.request.TokenEndpointAuthMethod = "private_key_jwt"
	_, err = registrar.Register(t.Context(), requestJWT)

	assert.NoError(t, err)
}

const sectorURI = "https://recipient.example/sector.json"

// Verifies: DH-2630
func TestARegistrationWithASectorIdentifierURIGetsTheDocumentAtEachRegistrationAndUpdate(t *testing.T) {
	j := validJWTs()
	j.ssa.SectorIdentifierURI = sectorURI
	registrar := newRegistrar(j)
	sectors := &sectorDocuments{redirectURIs: []string{"https://recipient.example/cb"}}
	registrar.Sectors = sectors

	registration, err := registrar.Register(t.Context(), requestJWT)
	require.NoError(t, err)

	j.request.ID, j.ssa.ID = "request-jti-2", "ssa-jti-2"
	_, err = registrar.Update(t.Context(), registration.ClientID, requestJWT)
	require.NoError(t, err)

	assert.Equal(t, []string{sectorURI, sectorURI}, sectors.requested)
}

// Verifies: DH-2630
func TestARegistrationWithNoSectorIdentifierURIGetsNoDocument(t *testing.T) {
	registrar := newRegistrar(validJWTs())
	sectors := &sectorDocuments{}
	registrar.Sectors = sectors

	_, err := registrar.Register(t.Context(), requestJWT)

	require.NoError(t, err)
	assert.Empty(t, sectors.requested)
}

// Verifies: DH-2630
func TestTheDataHolderGetsNoSectorIdentifierDocumentForARequestThatDoesNotVerify(t *testing.T) {
	j := validJWTs()
	j.ssa.SectorIdentifierURI = sectorURI
	j.requestErr = app.ErrInvalidJWT
	registrar := newRegistrar(j)
	sectors := &sectorDocuments{}
	registrar.Sectors = sectors

	_, err := registrar.Register(t.Context(), requestJWT)

	require.Error(t, err)
	assert.Empty(t, sectors.requested)
}

// Verifies: DH-2635, DH-2636
func TestARegistrationThatTheSectorIdentifierDocumentDoesNotPermitIsRefused(t *testing.T) {
	tests := map[string]struct {
		sectors     *sectorDocuments
		code        domain.RegistrationErrorCode
		description string
	}{
		"a document that is not available": {
			&sectorDocuments{
				err: app.ErrSectorDocumentNotAvailable,
			},
			domain.InvalidClientMetadata,
			"sector_identifier_uri",
		},
		"a document with no redirect URI of the registration": {
			&sectorDocuments{redirectURIs: []string{"https://recipient.example/other"}},
			domain.InvalidRedirectURI,
			"sector_identifier_uri",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			j := validJWTs()
			j.ssa.SectorIdentifierURI = sectorURI
			registrar := newRegistrar(j)
			registrar.Sectors = tc.sectors

			_, err := registrar.Register(t.Context(), requestJWT)
			requireRegistrationError(t, err, tc.code, tc.description)

			tc.sectors.redirectURIs, tc.sectors.err = []string{"https://recipient.example/cb"}, nil
			_, err = registrar.Register(t.Context(), requestJWT)
			assert.NoError(t, err, "the refused request did not use its jti")
		})
	}
}

func TestAFailureToGetTheSectorIdentifierDocumentThatIsNotTheFaultOfTheRequestIsNotARegistrationError(t *testing.T) {
	j := validJWTs()
	j.ssa.SectorIdentifierURI = sectorURI
	registrar := newRegistrar(j)
	registrar.Sectors = &sectorDocuments{err: errors.New("the context is cancelled")}

	_, err := registrar.Register(t.Context(), requestJWT)

	require.Error(t, err)

	var registrationErr domain.RegistrationError
	assert.NotErrorAs(t, err, &registrationErr)
}

// registeredClient gives a registrar with one registration, whose client ID is client-1.
func registeredClient(t *testing.T, j *jwts) app.Registrar {
	t.Helper()

	registrar := newRegistrar(j)
	_, err := registrar.Register(t.Context(), requestJWT)
	require.NoError(t, err)

	j.request.ID, j.ssa.ID = "request-jti-2", "ssa-jti-2"

	return registrar
}

func TestAnUpdateReplacesTheRegistrationAndKeepsTheClientID(t *testing.T) {
	j := validJWTs()
	registrar := registeredClient(t, j)
	j.request.GrantTypes = []string{"client_credentials", "authorization_code"}
	j.request.SoftwareStatement = "header.new-ssa.signature"
	registrar.Now = func() time.Time { return registrationTime.Add(30 * time.Second) }

	updated, err := registrar.Update(t.Context(), "client-1", requestJWT)

	require.NoError(t, err)
	assert.Equal(t, "client-1", updated.ClientID)
	assert.Equal(
		t,
		registrationTime.Unix(),
		updated.ClientIDIssuedAt,
		"the issue time of the client ID does not change",
	)
	assert.Equal(t, "header.new-ssa.signature", updated.SoftwareStatement)
	assert.Equal(t, []domain.GrantType{"client_credentials", "authorization_code"}, updated.GrantTypes)

	stored, err := registrar.Registrations.Registration(t.Context(), "client-1")
	require.NoError(t, err)
	assert.Equal(t, updated, stored)
}

func TestAnUpdateThatIsRefusedDoesNotChangeTheRegistration(t *testing.T) {
	tests := map[string]struct {
		change      func(*jwts)
		code        domain.RegistrationErrorCode
		description string
	}{
		"an SSA of a different software product": {
			func(j *jwts) { j.ssa.SoftwareID, j.request.Issuer = "software-2", "software-2" },
			domain.InvalidSoftwareStatement, "software_id",
		},
		"an SSA that does not verify": {
			func(j *jwts) { j.ssaErr = app.ErrInvalidJWT }, domain.InvalidSoftwareStatement, "signature",
		},
		"a request with a claim that is not valid": {
			func(j *jwts) { j.request.GrantTypes = []string{"implicit"} }, domain.InvalidClientMetadata, "grant_types",
		},
		"a jti that was used before": {
			func(j *jwts) { j.request.ID = "request-jti" }, domain.InvalidClientMetadata, "jti",
		},
		"the SSA of the registration": {
			func(j *jwts) { j.ssa.ID = "ssa-jti" }, domain.InvalidSoftwareStatement, "jti of the SSA",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			j := validJWTs()
			registrar := registeredClient(t, j)
			before, err := registrar.Registrations.Registration(t.Context(), "client-1")
			require.NoError(t, err)

			tc.change(j)

			_, err = registrar.Update(t.Context(), "client-1", requestJWT)

			requireRegistrationError(t, err, tc.code, tc.description)

			after, err := registrar.Registrations.Registration(t.Context(), "client-1")
			require.NoError(t, err)
			assert.Equal(t, before, after)
		})
	}
}

func TestAnUpdateWithAnSSAOfADifferentSoftwareProductDoesNotUseTheJTI(t *testing.T) {
	j := validJWTs()
	registrar := registeredClient(t, j)
	j.ssa.SoftwareID, j.request.Issuer = "software-2", "software-2"
	_, err := registrar.Update(t.Context(), "client-1", requestJWT)
	require.Error(t, err)

	j.ssa.SoftwareID, j.request.Issuer = "software-1", "software-1"
	_, err = registrar.Update(t.Context(), "client-1", requestJWT)

	assert.NoError(t, err)
}

// Verifies: DH-0634, DH-0675
func TestADeletionRemovesTheRegistrationAndANewRegistrationGetsANewClientID(t *testing.T) {
	j := validJWTs()
	registrar := registeredClient(t, j)

	require.NoError(t, registrar.Delete(t.Context(), "client-1"))

	_, err := registrar.Registrations.Registration(t.Context(), "client-1")
	require.ErrorIs(t, err, app.ErrRegistrationNotFound)

	j.ssa.ID = "ssa-jti-3"
	again, err := registrar.Register(t.Context(), requestJWT)
	require.NoError(t, err)
	assert.Equal(t, "client-2", again.ClientID)
}

// Verifies: DH-0676
func TestADeletionEndsTheArrangementsOfTheClientAtTheTimeOfTheRequest(t *testing.T) {
	registrar := registeredClient(t, validJWTs())
	arrangements := &clientArrangements{}
	registrar.Arrangements = arrangements

	require.NoError(t, registrar.Delete(t.Context(), "client-1"))

	assert.Equal(t, []string{"software-1/client-1 at " + registrationTime.Format(time.RFC3339)}, arrangements.ends)
}

func TestADeletionThatCannotEndTheArrangementsKeepsTheRegistration(t *testing.T) {
	registrar := registeredClient(t, validJWTs())
	arrangements := &clientArrangements{err: errors.New("database failure")}
	registrar.Arrangements = arrangements

	err := registrar.Delete(t.Context(), "client-1")

	require.ErrorIs(t, err, arrangements.err)

	_, err = registrar.Registrations.Registration(t.Context(), "client-1")
	assert.NoError(t, err, "a new DELETE can end the arrangements")
}

func TestAnUpdateOrADeletionOfAClientIDWithNoRegistrationIsNotFound(t *testing.T) {
	registrar := registeredClient(t, validJWTs())

	_, err := registrar.Update(t.Context(), "client-9", requestJWT)
	require.ErrorIs(t, err, app.ErrRegistrationNotFound)

	assert.ErrorIs(t, registrar.Delete(t.Context(), "client-9"), app.ErrRegistrationNotFound)
}

func TestAnUpdateAndADeletionReadTheStreamAgainAfterAVersionConflict(t *testing.T) {
	j := validJWTs()
	registrar := registeredClient(t, j)
	store := &conflictingStore{memoryStore: registrar.Store.(*memoryStore), conflicts: 2}
	registrar.Store = store

	_, err := registrar.Update(t.Context(), "client-1", requestJWT)
	require.NoError(t, err)

	store.conflicts = 2

	require.NoError(t, registrar.Delete(t.Context(), "client-1"))

	t.Run("not more than three times", func(t *testing.T) {
		j := validJWTs()
		registrar := registeredClient(t, j)
		registrar.Store = &conflictingStore{memoryStore: registrar.Store.(*memoryStore), conflicts: 3}

		err := registrar.Delete(t.Context(), "client-1")

		require.ErrorIs(t, err, eventsourcing.ErrVersionConflict)
	})
}
