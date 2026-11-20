package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

const issuer = "https://id.bank.example"

var now = time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

func validSSA() domain.SSA {
	return domain.SSA{
		Issuer:            "cdr-register",
		IssuedAt:          now.Add(-time.Minute).Unix(),
		ExpiresAt:         now.Add(9 * time.Minute).Unix(),
		ID:                "ssa-jti",
		LegalEntityID:     "legal-entity-1",
		LegalEntityName:   "Recipient Pty Ltd",
		OrgID:             "brand-1",
		OrgName:           "Recipient",
		ClientName:        "Recipient Web",
		ClientDescription: "A software product",
		ClientURI:         "https://recipient.example",
		RedirectURIs:      []string{"https://recipient.example/cb1", "https://recipient.example/cb2"},
		LogoURI:           "https://recipient.example/logo.png",
		JWKSURI:           "https://recipient.example/jwks",
		RevocationURI:     "https://recipient.example/revoke",
		RecipientBaseURI:  "https://recipient.example",
		SoftwareID:        "software-1",
		SoftwareRoles:     "data-recipient-software-product",
		Scope:             "openid profile bank:payees:read bank:accounts.basic:read cdr:registration",
	}
}

func validRequest() domain.RegistrationRequest {
	//nolint:gosec // G101 reads the protocol field token_endpoint_auth_method as a credential
	return domain.RegistrationRequest{
		Issuer:                         "software-1",
		IssuedAt:                       now.Unix(),
		ExpiresAt:                      now.Add(5 * time.Minute).Unix(),
		ID:                             "request-jti",
		Audience:                       []string{issuer},
		RedirectURIs:                   []string{"https://recipient.example/cb1"},
		TokenEndpointAuthMethod:        "private_key_jwt",
		TokenEndpointAuthSigningAlg:    "PS256",
		GrantTypes:                     []string{"client_credentials", "authorization_code", "refresh_token"},
		ResponseTypes:                  []string{"code"},
		ApplicationType:                "web",
		IDTokenSignedResponseAlg:       "PS256",
		AuthorizationSignedResponseAlg: "PS256",
		RequestObjectSigningAlg:        "ES256",
		SoftwareStatement:              "header.ssa.signature",
	}
}

func requireRegistrationError(t *testing.T, err error, code domain.RegistrationErrorCode, description string) {
	t.Helper()

	var registrationErr domain.RegistrationError
	require.ErrorAs(t, err, &registrationErr)
	assert.Equal(t, code, registrationErr.Code)
	assert.Contains(t, registrationErr.Description, description)
}

func TestAValidSSAIsAccepted(t *testing.T) {
	assert.NoError(t, validSSA().Validate(now))
}

// Verifies: DH-0615, DH-0616, DH-0617, DH-0618, DH-0619, DH-2631
func TestAnSSAWithAClaimThatIsNotValidIsRefused(t *testing.T) {
	tests := map[string]struct {
		change      func(*domain.SSA)
		description string
	}{
		"a different issuer":   {func(s *domain.SSA) { s.Issuer = "other-register" }, "iss"},
		"expired":              {func(s *domain.SSA) { s.ExpiresAt = now.Unix() }, "expired"},
		"no jti":               {func(s *domain.SSA) { s.ID = "" }, "jti"},
		"no iat":               {func(s *domain.SSA) { s.IssuedAt = 0 }, "iat"},
		"no org_id":            {func(s *domain.SSA) { s.OrgID = "" }, "org_id"},
		"no org_name":          {func(s *domain.SSA) { s.OrgName = "" }, "org_name"},
		"no client_name":       {func(s *domain.SSA) { s.ClientName = "" }, "client_name"},
		"no description":       {func(s *domain.SSA) { s.ClientDescription = "" }, "client_description"},
		"no client_uri":        {func(s *domain.SSA) { s.ClientURI = "" }, "client_uri"},
		"no redirect_uris":     {func(s *domain.SSA) { s.RedirectURIs = nil }, "redirect_uris"},
		"no logo_uri":          {func(s *domain.SSA) { s.LogoURI = "" }, "logo_uri"},
		"no jwks_uri":          {func(s *domain.SSA) { s.JWKSURI = "" }, "jwks_uri"},
		"a jwks_uri with http": {func(s *domain.SSA) { s.JWKSURI = "http://recipient.example/jwks" }, "jwks_uri"},
		"a sector_identifier_uri with http": {
			func(s *domain.SSA) { s.SectorIdentifierURI = "http://recipient.example/sector.json" },
			"sector_identifier_uri",
		},
		"no revocation_uri": {func(s *domain.SSA) { s.RevocationURI = "" }, "revocation_uri"},
		"no recipient base": {func(s *domain.SSA) { s.RecipientBaseURI = "" }, "recipient_base_uri"},
		"no software_id":    {func(s *domain.SSA) { s.SoftwareID = "" }, "software_id"},
		"different roles":   {func(s *domain.SSA) { s.SoftwareRoles = "data-holder" }, "software_roles"},
		"no scope":          {func(s *domain.SSA) { s.Scope = "" }, "scope"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ssa := validSSA()
			tc.change(&ssa)

			requireRegistrationError(t, ssa.Validate(now), domain.InvalidSoftwareStatement, tc.description)
		})
	}
}

// Verifies: DH-0635, DH-0636
func TestARegistrationHasTheValuesOfTheSSAAndTheRequest(t *testing.T) {
	ssa := validSSA()
	ssa.SectorIdentifierURI = "https://recipient.example/sector.json"
	ssa.TOSURI = "https://recipient.example/tos"
	ssa.PolicyURI = "https://recipient.example/policy"

	registration, err := domain.NewRegistration(ssa, validRequest(), issuer, now)

	require.NoError(t, err)
	//nolint:gosec // G101 reads the protocol field token_endpoint_auth_method as a credential
	assert.Equal(t, domain.Registration{
		ClientName:                     "Recipient Web",
		ClientDescription:              "A software product",
		ClientURI:                      "https://recipient.example",
		LegalEntityID:                  "legal-entity-1",
		LegalEntityName:                "Recipient Pty Ltd",
		OrgID:                          "brand-1",
		OrgName:                        "Recipient",
		RedirectURIs:                   []string{"https://recipient.example/cb1"},
		SectorIdentifierURI:            "https://recipient.example/sector.json",
		LogoURI:                        "https://recipient.example/logo.png",
		TOSURI:                         "https://recipient.example/tos",
		PolicyURI:                      "https://recipient.example/policy",
		JWKSURI:                        "https://recipient.example/jwks",
		RevocationURI:                  "https://recipient.example/revoke",
		RecipientBaseURI:               "https://recipient.example",
		TokenEndpointAuthMethod:        "private_key_jwt",
		TokenEndpointAuthSigningAlg:    domain.PS256,
		GrantTypes:                     []domain.GrantType{"client_credentials", "authorization_code", "refresh_token"},
		ResponseTypes:                  []string{"code"},
		ApplicationType:                "web",
		IDTokenSignedResponseAlg:       domain.PS256,
		AuthorizationSignedResponseAlg: domain.PS256,
		RequestObjectSigningAlg:        domain.ES256,
		SoftwareStatement:              "header.ssa.signature",
		SoftwareID:                     "software-1",
		SoftwareRoles:                  "data-recipient-software-product",
		Scope:                          "openid profile bank:accounts.basic:read cdr:registration",
	}, registration)
}

// Verifies: DH-0640, DH-0641
func TestARegistrationHasTheDefaultsForClaimsThatTheRequestDoesNotHave(t *testing.T) {
	request := validRequest()
	request.RedirectURIs = nil
	request.RequestObjectSigningAlg = ""
	request.ApplicationType = ""

	registration, err := domain.NewRegistration(validSSA(), request, issuer, now)

	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{"https://recipient.example/cb1", "https://recipient.example/cb2"},
		registration.RedirectURIs,
	)
	assert.Equal(t, domain.PS256, registration.RequestObjectSigningAlg)
	assert.Equal(t, "web", registration.ApplicationType)
}

// Verifies: DH-0630
func TestARegistrationHasOnlyTheResponseTypeCode(t *testing.T) {
	request := validRequest()
	request.ResponseTypes = []string{"code id_token", "code"}

	registration, err := domain.NewRegistration(validSSA(), request, issuer, now)

	require.NoError(t, err)
	assert.Equal(t, []string{"code"}, registration.ResponseTypes)
}

// Verifies: DH-2634, DH-2636
func TestARegistrationWithARedirectURIThatTheSectorIdentifierDocumentDoesNotHaveIsRefused(t *testing.T) {
	registration := domain.Registration{
		RedirectURIs: []string{"https://recipient.example/cb", "https://recipient.example/other"},
	}

	t.Run("each redirect URI is in the document", func(t *testing.T) {
		published := []string{
			"https://recipient.example/other",
			"https://recipient.example/cb",
			"https://recipient.example/more",
		}

		assert.NoError(t, registration.CheckSectorIdentifierDocument(published))
	})

	t.Run("a redirect URI is not in the document", func(t *testing.T) {
		err := registration.CheckSectorIdentifierDocument([]string{"https://recipient.example/cb"})

		requireRegistrationError(t, err, domain.InvalidRedirectURI, "sector_identifier_uri")
	})
}

// Verifies: DH-0631, DH-0652
func TestARequestWithARedirectURIThatTheSSADoesNotHaveIsRefused(t *testing.T) {
	request := validRequest()
	request.RedirectURIs = []string{"https://recipient.example/cb1", "https://other.example/cb"}

	_, err := domain.NewRegistration(validSSA(), request, issuer, now)

	requireRegistrationError(t, err, domain.InvalidRedirectURI, "redirect_uris")
}

// Verifies: DH-0604, DH-0605, DH-0606, DH-0607, DH-0624, DH-0625, DH-0626, DH-0627, DH-0628, DH-0629
func TestARequestWithAClaimThatIsNotValidIsRefused(t *testing.T) {
	tests := map[string]struct {
		change      func(*domain.RegistrationRequest)
		description string
	}{
		"an iss that is not the software_id": {func(r *domain.RegistrationRequest) { r.Issuer = "software-2" }, "iss"},
		"an aud that is not the issuer": {
			func(r *domain.RegistrationRequest) { r.Audience = []string{"https://other.example"} }, "aud",
		},
		"no aud":  {func(r *domain.RegistrationRequest) { r.Audience = nil }, "aud"},
		"no iat":  {func(r *domain.RegistrationRequest) { r.IssuedAt = 0 }, "iat"},
		"no jti":  {func(r *domain.RegistrationRequest) { r.ID = "" }, "jti"},
		"expired": {func(r *domain.RegistrationRequest) { r.ExpiresAt = now.Unix() }, "expired"},
		"an exp of more than 10 minutes": {
			func(r *domain.RegistrationRequest) { r.ExpiresAt = now.Add(10*time.Minute + time.Second).Unix() }, "exp",
		},
		"a different authentication method": {
			func(r *domain.RegistrationRequest) { r.TokenEndpointAuthMethod = "client_secret_basic" },
			"token_endpoint_auth_method",
		},
		"a token algorithm that is not permitted": {
			func(r *domain.RegistrationRequest) { r.TokenEndpointAuthSigningAlg = "RS256" },
			"token_endpoint_auth_signing_alg",
		},
		"no token algorithm": {
			func(r *domain.RegistrationRequest) { r.TokenEndpointAuthSigningAlg = "" },
			"token_endpoint_auth_signing_alg",
		},
		"a request object algorithm that is not permitted": {
			func(r *domain.RegistrationRequest) { r.RequestObjectSigningAlg = "none" }, "request_object_signing_alg",
		},
		"an ID token algorithm that the data holder does not sign with": {
			func(r *domain.RegistrationRequest) { r.IDTokenSignedResponseAlg = "ES256" },
			"id_token_signed_response_alg",
		},
		"no authorisation response algorithm": {
			func(r *domain.RegistrationRequest) { r.AuthorizationSignedResponseAlg = "" },
			"authorization_signed_response_alg",
		},
		"a different application type": {
			func(r *domain.RegistrationRequest) { r.ApplicationType = "native" }, "application_type",
		},
		"no grant types": {func(r *domain.RegistrationRequest) { r.GrantTypes = nil }, "grant_types"},
		"a grant type that is not supported": {
			func(r *domain.RegistrationRequest) { r.GrantTypes = []string{"client_credentials", "implicit"} },
			"grant_types",
		},
		"no response type code": {
			func(r *domain.RegistrationRequest) { r.ResponseTypes = []string{"code id_token"} }, "response_types",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			request := validRequest()
			tc.change(&request)

			_, err := domain.NewRegistration(validSSA(), request, issuer, now)

			requireRegistrationError(t, err, domain.InvalidClientMetadata, tc.description)
		})
	}
}

// Verifies: DH-0607
func TestAnExpOfExactly10MinutesIsAccepted(t *testing.T) {
	request := validRequest()
	request.ExpiresAt = now.Add(10 * time.Minute).Unix()

	_, err := domain.NewRegistration(validSSA(), request, issuer, now)

	assert.NoError(t, err)
}

// Verifies: DH-0632, DH-0651
func TestASoftwareProductRegistersOneTime(t *testing.T) {
	registration, err := domain.NewRegistration(validSSA(), validRequest(), issuer, now)
	require.NoError(t, err)

	registration.ClientID = "client-1"
	registration.ClientIDIssuedAt = now.Unix()

	events, err := domain.HandleClientRegistration(
		domain.ClientRegistrationState{},
		domain.RegisterClient{Registration: registration},
	)

	require.NoError(t, err)
	require.Equal(t, []domain.ClientRegistrationEvent{domain.ClientRegistered{Registration: registration}}, events)

	state := domain.ApplyClientRegistration(domain.ClientRegistrationState{}, events[0])
	assert.True(t, state.Registered)
	assert.Equal(t, registration, state.Registration)

	_, err = domain.HandleClientRegistration(state, domain.RegisterClient{Registration: registration})

	requireRegistrationError(t, err, domain.InvalidSoftwareStatement, "software_id")
}

func registeredState(t *testing.T) domain.ClientRegistrationState {
	t.Helper()

	registration, err := domain.NewRegistration(validSSA(), validRequest(), issuer, now)
	require.NoError(t, err)

	registration.ClientID = "client-1"
	registration.ClientIDIssuedAt = now.Unix()

	return domain.ApplyClientRegistration(
		domain.ClientRegistrationState{},
		domain.ClientRegistered{Registration: registration},
	)
}

// Verifies: DH-0671
func TestAnUpdateReplacesAllValuesButNotTheClientIDAndItsIssueTime(t *testing.T) {
	state := registeredState(t)
	request := validRequest()
	request.GrantTypes = []string{"client_credentials"}
	request.RedirectURIs = []string{"https://recipient.example/cb2"}
	replacement, err := domain.NewRegistration(validSSA(), request, issuer, now)
	require.NoError(t, err)

	events, err := domain.HandleClientRegistration(
		state,
		domain.UpdateClient{ClientID: "client-1", Registration: replacement},
	)

	require.NoError(t, err)
	require.Len(t, events, 1)

	updated := domain.ApplyClientRegistration(state, events[0])
	assert.True(t, updated.Registered)
	assert.Equal(t, "client-1", updated.Registration.ClientID)
	assert.Equal(t, now.Unix(), updated.Registration.ClientIDIssuedAt)
	assert.Equal(t, []domain.GrantType{domain.GrantClientCredentials}, updated.Registration.GrantTypes)
	assert.Equal(t, []string{"https://recipient.example/cb2"}, updated.Registration.RedirectURIs)
}

// Verifies: DH-0670
func TestAnUpdateWithAnSSAOfADifferentSoftwareProductIsRefused(t *testing.T) {
	ssa := validSSA()
	ssa.SoftwareID = "software-2"
	request := validRequest()
	request.Issuer = "software-2"
	replacement, err := domain.NewRegistration(ssa, request, issuer, now)
	require.NoError(t, err)

	_, err = domain.HandleClientRegistration(
		registeredState(t),
		domain.UpdateClient{ClientID: "client-1", Registration: replacement},
	)

	requireRegistrationError(t, err, domain.InvalidSoftwareStatement, "software_id")
}

// Verifies: DH-0675
func TestADeletionRemovesTheRegistrationAndTheSoftwareProductCanRegisterAgain(t *testing.T) {
	state := registeredState(t)

	events, err := domain.HandleClientRegistration(state, domain.DeleteClient{ClientID: "client-1"})

	require.NoError(t, err)
	require.Equal(t, []domain.ClientRegistrationEvent{domain.ClientDeleted{ClientID: "client-1"}}, events)

	deleted := domain.ApplyClientRegistration(state, events[0])
	assert.False(t, deleted.Registered)

	again := state.Registration
	again.ClientID = "client-2"
	events, err = domain.HandleClientRegistration(deleted, domain.RegisterClient{Registration: again})

	require.NoError(t, err)
	assert.Equal(t, "client-2", domain.ApplyClientRegistration(deleted, events[0]).Registration.ClientID)
}

func TestAnUpdateOrADeletionNeedsTheRegistrationOfTheClientID(t *testing.T) {
	states := map[string]domain.ClientRegistrationState{
		"no registration":                       {},
		"a registration with another client ID": registeredState(t),
	}
	commands := map[string]func(clientID string) domain.ClientRegistrationCommand{
		"update": func(id string) domain.ClientRegistrationCommand { return domain.UpdateClient{ClientID: id} },
		"delete": func(id string) domain.ClientRegistrationCommand { return domain.DeleteClient{ClientID: id} },
	}

	for stateName, state := range states {
		for commandName, command := range commands {
			t.Run(commandName+" with "+stateName, func(t *testing.T) {
				_, err := domain.HandleClientRegistration(state, command("client-9"))

				assert.ErrorIs(t, err, domain.ErrNoRegistration)
			})
		}
	}
}
