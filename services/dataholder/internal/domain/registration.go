package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	ssaIssuer          = "cdr-register"
	ssaSoftwareRoles   = "data-recipient-software-product"
	maxRequestLifetime = 10 * time.Minute

	authMethodPrivateKeyJWT = "private_key_jwt"
	applicationTypeWeb      = "web"
	responseTypeCode        = "code"
)

type SigningAlg string

const (
	PS256 SigningAlg = "PS256"
	ES256 SigningAlg = "ES256"
)

type GrantType string

const (
	GrantClientCredentials GrantType = "client_credentials"
	GrantAuthorizationCode GrantType = "authorization_code"
	GrantRefreshToken      GrantType = "refresh_token"
)

// RegistrationErrorCode is an error value of RFC 7591, section 3.2.2.
type RegistrationErrorCode string

const (
	InvalidRedirectURI          RegistrationErrorCode = "invalid_redirect_uri"
	InvalidClientMetadata       RegistrationErrorCode = "invalid_client_metadata"
	InvalidSoftwareStatement    RegistrationErrorCode = "invalid_software_statement"
	UnapprovedSoftwareStatement RegistrationErrorCode = "unapproved_software_statement"
)

// RegistrationError goes to the software product, so its description names a claim or a rule and never a value.
type RegistrationError struct {
	Code        RegistrationErrorCode
	Description string
}

func (e RegistrationError) Error() string {
	return string(e.Code) + ": " + e.Description
}

func invalidSSA(format string, args ...any) error {
	return RegistrationError{Code: InvalidSoftwareStatement, Description: fmt.Sprintf(format, args...)}
}

func invalidMetadata(format string, args ...any) error {
	return RegistrationError{Code: InvalidClientMetadata, Description: fmt.Sprintf(format, args...)}
}

// SSA has the claims of a Software Statement Assertion.
type SSA struct {
	Issuer              string   `json:"iss"`
	IssuedAt            int64    `json:"iat"`
	ExpiresAt           int64    `json:"exp"`
	ID                  string   `json:"jti"`
	LegalEntityID       string   `json:"legal_entity_id"`
	LegalEntityName     string   `json:"legal_entity_name"`
	OrgID               string   `json:"org_id"`
	OrgName             string   `json:"org_name"`
	ClientName          string   `json:"client_name"`
	ClientDescription   string   `json:"client_description"`
	ClientURI           string   `json:"client_uri"`
	RedirectURIs        []string `json:"redirect_uris"`
	SectorIdentifierURI string   `json:"sector_identifier_uri"`
	LogoURI             string   `json:"logo_uri"`
	TOSURI              string   `json:"tos_uri"`
	PolicyURI           string   `json:"policy_uri"`
	JWKSURI             string   `json:"jwks_uri"`
	RevocationURI       string   `json:"revocation_uri"`
	RecipientBaseURI    string   `json:"recipient_base_uri"`
	SoftwareID          string   `json:"software_id"`
	SoftwareRoles       string   `json:"software_roles"`
	Scope               string   `json:"scope"`
}

// Validate does not check the signature.
func (s SSA) Validate(now time.Time) error {
	if s.Issuer != ssaIssuer {
		return invalidSSA("the SSA iss is not %s", ssaIssuer)
	}

	if s.IssuedAt == 0 {
		return invalidSSA("the SSA has no iat")
	}

	if !now.Before(time.Unix(s.ExpiresAt, 0)) {
		return invalidSSA("the SSA is expired")
	}

	required := []struct{ claim, value string }{
		{"jti", s.ID},
		{"org_id", s.OrgID},
		{"org_name", s.OrgName},
		{"client_name", s.ClientName},
		{"client_description", s.ClientDescription},
		{"client_uri", s.ClientURI},
		{"logo_uri", s.LogoURI},
		{"jwks_uri", s.JWKSURI},
		{"revocation_uri", s.RevocationURI},
		{"recipient_base_uri", s.RecipientBaseURI},
		{"software_id", s.SoftwareID},
		{"scope", s.Scope},
	}
	for _, r := range required {
		if r.value == "" {
			return invalidSSA("the SSA has no %s", r.claim)
		}
	}

	if len(s.RedirectURIs) == 0 {
		return invalidSSA("the SSA has no redirect_uris")
	}

	if !strings.HasPrefix(s.JWKSURI, "https://") {
		return invalidSSA("the SSA jwks_uri is not an https URL")
	}

	if s.SectorIdentifierURI != "" && !strings.HasPrefix(s.SectorIdentifierURI, "https://") {
		return invalidSSA("the SSA sector_identifier_uri is not an https URL")
	}

	if s.SoftwareRoles != ssaSoftwareRoles {
		return invalidSSA("the SSA software_roles is not %s", ssaSoftwareRoles)
	}

	return nil
}

// RegistrationRequest has the claims of a registration request JWT that the data holder supports.
type RegistrationRequest struct {
	Issuer                         string   `json:"iss"`
	IssuedAt                       int64    `json:"iat"`
	ExpiresAt                      int64    `json:"exp"`
	ID                             string   `json:"jti"`
	Audience                       []string `json:"aud"`
	RedirectURIs                   []string `json:"redirect_uris"`
	TokenEndpointAuthMethod        string   `json:"token_endpoint_auth_method"`
	TokenEndpointAuthSigningAlg    string   `json:"token_endpoint_auth_signing_alg"`
	GrantTypes                     []string `json:"grant_types"`
	ResponseTypes                  []string `json:"response_types"`
	ApplicationType                string   `json:"application_type"`
	IDTokenSignedResponseAlg       string   `json:"id_token_signed_response_alg"`
	AuthorizationSignedResponseAlg string   `json:"authorization_signed_response_alg"`
	RequestObjectSigningAlg        string   `json:"request_object_signing_alg"`
	SoftwareStatement              string   `json:"software_statement"`
}

// Registration is the client metadata that the data holder holds for one software product.
type Registration struct {
	ClientID                       string      `json:"client_id"`
	ClientIDIssuedAt               int64       `json:"client_id_issued_at"`
	ClientName                     string      `json:"client_name"`
	ClientDescription              string      `json:"client_description"`
	ClientURI                      string      `json:"client_uri"`
	LegalEntityID                  string      `json:"legal_entity_id,omitempty"`
	LegalEntityName                string      `json:"legal_entity_name,omitempty"`
	OrgID                          string      `json:"org_id"`
	OrgName                        string      `json:"org_name"`
	RedirectURIs                   []string    `json:"redirect_uris"`
	SectorIdentifierURI            string      `json:"sector_identifier_uri,omitempty"`
	LogoURI                        string      `json:"logo_uri"`
	TOSURI                         string      `json:"tos_uri,omitempty"`
	PolicyURI                      string      `json:"policy_uri,omitempty"`
	JWKSURI                        string      `json:"jwks_uri"`
	RevocationURI                  string      `json:"revocation_uri"`
	RecipientBaseURI               string      `json:"recipient_base_uri"`
	TokenEndpointAuthMethod        string      `json:"token_endpoint_auth_method"`
	TokenEndpointAuthSigningAlg    SigningAlg  `json:"token_endpoint_auth_signing_alg"`
	GrantTypes                     []GrantType `json:"grant_types"`
	ResponseTypes                  []string    `json:"response_types"`
	ApplicationType                string      `json:"application_type"`
	IDTokenSignedResponseAlg       SigningAlg  `json:"id_token_signed_response_alg"`
	AuthorizationSignedResponseAlg SigningAlg  `json:"authorization_signed_response_alg"`
	RequestObjectSigningAlg        SigningAlg  `json:"request_object_signing_alg"`
	SoftwareStatement              string      `json:"software_statement"`
	SoftwareID                     string      `json:"software_id"`
	SoftwareRoles                  string      `json:"software_roles"`
	Scope                          string      `json:"scope"`
}

// NewRegistration gives a registration with no client ID; ssa must be valid, and issuer is the issuer of the data
// holder.
func NewRegistration(ssa SSA, request RegistrationRequest, issuer string, now time.Time) (Registration, error) {
	if err := request.validateJWT(ssa.SoftwareID, issuer, now); err != nil {
		return Registration{}, err
	}

	redirectURIs := request.RedirectURIs
	if len(redirectURIs) == 0 {
		redirectURIs = ssa.RedirectURIs
	}

	for _, uri := range redirectURIs {
		if !slices.Contains(ssa.RedirectURIs, uri) {
			return Registration{}, RegistrationError{
				Code:        InvalidRedirectURI,
				Description: "a value of redirect_uris is not in the redirect_uris of the SSA",
			}
		}
	}

	metadata, err := request.clientMetadata()
	if err != nil {
		return Registration{}, err
	}

	metadata.ClientName = ssa.ClientName
	metadata.ClientDescription = ssa.ClientDescription
	metadata.ClientURI = ssa.ClientURI
	metadata.LegalEntityID = ssa.LegalEntityID
	metadata.LegalEntityName = ssa.LegalEntityName
	metadata.OrgID = ssa.OrgID
	metadata.OrgName = ssa.OrgName
	metadata.RedirectURIs = redirectURIs
	metadata.SectorIdentifierURI = ssa.SectorIdentifierURI
	metadata.LogoURI = ssa.LogoURI
	metadata.TOSURI = ssa.TOSURI
	metadata.PolicyURI = ssa.PolicyURI
	metadata.JWKSURI = ssa.JWKSURI
	metadata.RevocationURI = ssa.RevocationURI
	metadata.RecipientBaseURI = ssa.RecipientBaseURI
	metadata.SoftwareStatement = request.SoftwareStatement
	metadata.SoftwareID = ssa.SoftwareID
	metadata.SoftwareRoles = ssa.SoftwareRoles
	metadata.Scope = supportedScopesOf(ssa.Scope)

	return metadata, nil
}

func (r Registration) CheckSectorIdentifierDocument(published []string) error {
	for _, uri := range r.RedirectURIs {
		if !slices.Contains(published, uri) {
			return RegistrationError{
				Code:        InvalidRedirectURI,
				Description: "a value of redirect_uris is not in the document at the sector_identifier_uri",
			}
		}
	}

	return nil
}

func (r RegistrationRequest) validateJWT(softwareID, issuer string, now time.Time) error {
	if r.Issuer != softwareID {
		return invalidMetadata("the request iss is not the software_id of the SSA")
	}

	if !slices.Contains(r.Audience, issuer) {
		return invalidMetadata("the request aud is not the issuer of the data holder")
	}

	if r.IssuedAt == 0 {
		return invalidMetadata("the request has no iat")
	}

	if r.ID == "" {
		return invalidMetadata("the request has no jti")
	}

	expiry := time.Unix(r.ExpiresAt, 0)
	if !now.Before(expiry) {
		return invalidMetadata("the request is expired")
	}

	if expiry.After(now.Add(maxRequestLifetime)) {
		return invalidMetadata(
			"the request exp is more than %d minutes after the time of the request",
			int(maxRequestLifetime.Minutes()),
		)
	}

	return nil
}

// clientMetadata gives the part of a registration that comes from the request.
func (r RegistrationRequest) clientMetadata() (Registration, error) {
	if r.TokenEndpointAuthMethod != authMethodPrivateKeyJWT {
		return Registration{}, invalidMetadata("token_endpoint_auth_method is not %s", authMethodPrivateKeyJWT)
	}

	tokenAlg, err := signingAlg("token_endpoint_auth_signing_alg", r.TokenEndpointAuthSigningAlg, PS256, ES256)
	if err != nil {
		return Registration{}, err
	}

	requestObjectAlg := PS256
	if r.RequestObjectSigningAlg != "" {
		if requestObjectAlg, err = signingAlg(
			"request_object_signing_alg", r.RequestObjectSigningAlg, PS256, ES256,
		); err != nil {
			return Registration{}, err
		}
	}

	idTokenAlg, err := signingAlg("id_token_signed_response_alg", r.IDTokenSignedResponseAlg, PS256)
	if err != nil {
		return Registration{}, err
	}

	authorisationAlg, err := signingAlg("authorization_signed_response_alg", r.AuthorizationSignedResponseAlg, PS256)
	if err != nil {
		return Registration{}, err
	}

	if r.ApplicationType != "" && r.ApplicationType != applicationTypeWeb {
		return Registration{}, invalidMetadata("application_type is not %s", applicationTypeWeb)
	}

	grantTypes, err := grantTypes(r.GrantTypes)
	if err != nil {
		return Registration{}, err
	}

	if !slices.Contains(r.ResponseTypes, responseTypeCode) {
		return Registration{}, invalidMetadata("response_types does not have %s", responseTypeCode)
	}

	return Registration{
		TokenEndpointAuthMethod:        authMethodPrivateKeyJWT,
		TokenEndpointAuthSigningAlg:    tokenAlg,
		GrantTypes:                     grantTypes,
		ResponseTypes:                  []string{responseTypeCode},
		ApplicationType:                applicationTypeWeb,
		IDTokenSignedResponseAlg:       idTokenAlg,
		AuthorizationSignedResponseAlg: authorisationAlg,
		RequestObjectSigningAlg:        requestObjectAlg,
	}, nil
}

func signingAlg(claim, value string, permitted ...SigningAlg) (SigningAlg, error) {
	if !slices.Contains(permitted, SigningAlg(value)) {
		return "", invalidMetadata("%s is not a signing algorithm that the data holder supports", claim)
	}

	return SigningAlg(value), nil
}

func grantTypes(values []string) ([]GrantType, error) {
	if len(values) == 0 {
		return nil, invalidMetadata("the request has no grant_types")
	}

	supported := []GrantType{GrantClientCredentials, GrantAuthorizationCode, GrantRefreshToken}
	grants := make([]GrantType, 0, len(values))

	for _, v := range values {
		if !slices.Contains(supported, GrantType(v)) {
			return nil, invalidMetadata("a value of grant_types is not supported")
		}

		grants = append(grants, GrantType(v))
	}

	return grants, nil
}

// supportedScopesOf keeps the sequence of the SSA.
func supportedScopesOf(ssaScope string) string {
	supported := SupportedScopes()

	var scopes []string

	for scope := range strings.FieldsSeq(ssaScope) {
		if slices.Contains(supported, scope) {
			scopes = append(scopes, scope)
		}
	}

	return strings.Join(scopes, " ")
}

// ClientRegistrationState is the state of the registration of one software product.
type ClientRegistrationState struct {
	Registered bool
	// Registration is valid only when Registered is true.
	Registration Registration
}

// ErrNoRegistration means that the software product has no registration with the client ID of the command.
var ErrNoRegistration = errors.New("the software product has no registration with this client ID")

type ClientRegistrationCommand interface{ isClientRegistrationCommand() }

type RegisterClient struct {
	// Registration has the client ID that the data holder issues.
	Registration Registration
}

func (RegisterClient) isClientRegistrationCommand() {}

// UpdateClient replaces all values of the registration, but not its client ID and the time of its issue.
type UpdateClient struct {
	ClientID     string
	Registration Registration
}

func (UpdateClient) isClientRegistrationCommand() {}

type DeleteClient struct {
	ClientID string
}

func (DeleteClient) isClientRegistrationCommand() {}

type ClientRegistrationEvent interface{ isClientRegistrationEvent() }

type ClientRegistered struct {
	Registration Registration `json:"registration"`
}

func (ClientRegistered) isClientRegistrationEvent() {}

type ClientUpdated struct {
	Registration Registration `json:"registration"`
}

func (ClientUpdated) isClientRegistrationEvent() {}

type ClientDeleted struct {
	ClientID string `json:"client_id"`
}

func (ClientDeleted) isClientRegistrationEvent() {}

func HandleClientRegistration(
	state ClientRegistrationState,
	command ClientRegistrationCommand,
) ([]ClientRegistrationEvent, error) {
	switch c := command.(type) {
	case RegisterClient:
		if state.Registered {
			return nil, invalidSSA("the software_id of the SSA has a registration")
		}

		return []ClientRegistrationEvent{ClientRegistered(c)}, nil
	case UpdateClient:
		if !state.has(c.ClientID) {
			return nil, ErrNoRegistration
		}

		if c.Registration.SoftwareID != state.Registration.SoftwareID {
			return nil, invalidSSA("the software_id of the SSA is not the software_id of the registration")
		}

		updated := c.Registration
		updated.ClientID = state.Registration.ClientID
		updated.ClientIDIssuedAt = state.Registration.ClientIDIssuedAt

		return []ClientRegistrationEvent{ClientUpdated{Registration: updated}}, nil
	case DeleteClient:
		if !state.has(c.ClientID) {
			return nil, ErrNoRegistration
		}

		return []ClientRegistrationEvent{ClientDeleted(c)}, nil
	default:
		return nil, fmt.Errorf("the command %T is not a client registration command", command)
	}
}

func (s ClientRegistrationState) has(clientID string) bool {
	return s.Registered && s.Registration.ClientID == clientID
}

func ApplyClientRegistration(state ClientRegistrationState, event ClientRegistrationEvent) ClientRegistrationState {
	switch e := event.(type) {
	case ClientRegistered:
		return ClientRegistrationState{Registered: true, Registration: e.Registration}
	case ClientUpdated:
		return ClientRegistrationState{Registered: true, Registration: e.Registration}
	case ClientDeleted:
		return ClientRegistrationState{}
	default:
		return state
	}
}
