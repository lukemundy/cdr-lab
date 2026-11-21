package authserver

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/luikyv/go-oidc/pkg/provider"

	"github.com/lukemundy/cdr-lab/pkg/authserver/internal/cdsrules"
)

// The paths are the default paths of the OpenID Provider library, so that the library can serve them.
const (
	dataHolderConfigPath                = "/.well-known/openid-configuration"
	dataHolderJWKSPath                  = "/jwks"
	dataHolderAuthorisationPath         = "/authorize"
	dataHolderRegistrationPath          = "/register"
	dataHolderTokenPath                 = "/token"
	dataHolderUserInfoPath              = "/userinfo"
	dataHolderIntrospectionPath         = "/introspect"
	dataHolderRevocationPath            = "/revoke"
	dataHolderPARPath                   = "/par"
	dataHolderArrangementRevocationPath = "/arrangements/revoke"
)

// RegistrationScope is the scope of an access token for registration management.
const RegistrationScope = cdsrules.RegistrationScope

// AccessTokenLifetime is also the sharing of a once-off arrangement, so that the data recipient can collect the data
// with its only access token.
const AccessTokenLifetime = 5 * time.Minute

const (
	dataHolderTokenLifetimeSeconds = int(AccessTokenLifetime / time.Second)
	// A JTIStore must keep a JWT ID for longer than the 60 minutes of a request object plus clockLeewaySeconds.
	dataHolderAssertionLifetimeSeconds = 600
	// The CDS permits 10 to 90 seconds.
	dataHolderRequestURILifetimeSeconds = 60
	dataHolderCodeLifetimeSeconds       = 60
	// A session is some kilobytes, so the sessions use some tens of megabytes at most.
	dataHolderSessionLimit = 10_000
)

// DataHolderClient is a registered software product.
type DataHolderClient struct {
	ID      string
	JWKSURI string
	// Scopes, GrantTypes and RedirectURIs are those of the registration.
	Scopes       []string
	GrantTypes   []string
	RedirectURIs []string
	// RequestObjectSigningAlg is the only algorithm that the client signs a request object with.
	RequestObjectSigningAlg string
	// TokenEndpointAuthSigningAlg is the only algorithm that the client signs a client assertion with.
	TokenEndpointAuthSigningAlg string
}

type DataHolderClientStore interface {
	// Client gives ErrClientNotFound when the client does not exist.
	Client(ctx context.Context, id string) (DataHolderClient, error)
}

type DataHolderConfig struct {
	// Issuer is also the base URL of the TLS endpoints.
	Issuer string
	// SecureBaseURL is the base URL of the mTLS endpoints.
	SecureBaseURL string
	Scopes        []string
	Clients       DataHolderClientStore
	JTIs          JTIStore
	// SigningKeys gives the keys of the data holder. The first key signs. The JWKS publishes all keys.
	SigningKeys func(context.Context) ([]SigningKey, error)
	// JWKSTransport is optional: nil is http.DefaultTransport.
	JWKSTransport http.RoundTripper
	// Pages serves the pages on which a consumer authenticates and gives consent.
	Pages ConsumerPages
	// Grants keeps the authorisation code and the refresh token of each consent.
	Grants         GrantStore
	RevokedTokens  RevokedTokenStore
	Arrangements   ArrangementRevoker
	ClientStatuses ClientStatuses
}

type ClientStatuses interface {
	Active(ctx context.Context, clientID string) (bool, error)
}

type DataHolder struct {
	cfg      DataHolderConfig
	provider http.Handler
	sessions *sessionStore
	grants   libraryGrants
	now      func() time.Time
}

func NewDataHolder(cfg DataHolderConfig) (*DataHolder, error) {
	if cfg.Issuer == "" || cfg.SecureBaseURL == "" || len(cfg.Scopes) == 0 || cfg.Clients == nil ||
		cfg.JTIs == nil || cfg.SigningKeys == nil || cfg.Pages == nil ||
		cfg.Grants == nil || cfg.RevokedTokens == nil ||
		cfg.Arrangements == nil || cfg.ClientStatuses == nil {
		return nil, errors.New("authserver: all fields of DataHolderConfig except JWKSTransport are required")
	}

	cfg.Issuer = strings.TrimSuffix(cfg.Issuer, "/")
	cfg.SecureBaseURL = strings.TrimSuffix(cfg.SecureBaseURL, "/")

	jwksClient := NewJWKSClient(cfg.JWKSTransport)

	scopes := make([]goidc.Scope, 0, len(cfg.Scopes))
	for _, scope := range cfg.Scopes {
		scopes = append(scopes, goidc.NewScope(scope))
	}

	clientAlgs := []goidc.SignatureAlgorithm{goidc.SigAlgPS256, goidc.SigAlgES256}
	ownAlgs := []goidc.SignatureAlgorithm{goidc.SigAlgPS256}
	d := &DataHolder{cfg: cfg, grants: libraryGrants{cfg.Grants}, now: time.Now}
	sessions := newSessionStore(dataHolderSessionLimit, func() time.Time { return d.now() })
	d.sessions = sessions
	authorisations := flows{sessionStore: sessions, libraryGrants: d.grants, requireActive: d.requireActive}

	op, err := provider.New(
		provider.Config{
			Issuer:      cfg.Issuer,
			Manager:     d.grants,
			JWKS:        privateJWKS(cfg.SigningKeys),
			IDTokenAlgs: ownAlgs,
		},
		provider.WithProfile(goidc.ProfileFAPI1),
		provider.WithScopes(scopes...),
		provider.WithOpenIDScopeRequired(),
		provider.WithACRs(cdsrules.SingleFactorACR),
		provider.WithMTLS(provider.MTLSConfig{Host: cfg.SecureBaseURL, ClientCert: clientCertFromContext},
			provider.WithMTLSTokenBinding(provider.WithMTLSTokenBindingRequired())),
		provider.WithPrivateKeyJWTAuthn(clientAlgs...),
		provider.WithClientCredentialsGrant(),
		provider.WithAuthCodeGrant(
			provider.AuthCodeGrantConfig{
				Manager:       authorisations,
				ResponseTypes: []goidc.ResponseType{goidc.ResponseTypeCode},
			},
			provider.WithAuthCodeLifetime(dataHolderCodeLifetimeSeconds),
			provider.WithJAR(clientAlgs, provider.WithJARRequired()),
			provider.WithPAR(
				authorisations,
				provider.WithPARRequired(),
				provider.WithPARLifetime(dataHolderRequestURILifetimeSeconds),
				provider.WithPARSessionHandler(
					func(ctx context.Context, as *goidc.AuthnSession, c *goidc.Client) error {
						if err := d.requireActive(ctx, c.ID); err != nil {
							return err
						}

						return cdsrules.PushedRequest(ctx, as, c)
					},
				),
			),
			provider.WithPKCE(
				[]goidc.CodeChallengeMethod{goidc.CodeChallengeMethodSHA256},
				provider.WithPKCERequired(),
			),
			provider.WithJARM(ownAlgs),
			provider.WithClaimsParameter(),
			provider.WithAuthPolicies(consumerPolicy(cfg.Pages, d.requireActive)),
		),
		// The CDS forbids a new refresh token at a refresh, so the grant has no rotation option.
		provider.WithRefreshTokenGrant(
			cdsrules.RefreshWithAnyCertificate(d.grants, clientCertFromContext),
			provider.WithRefreshTokenShouldIssue(cdsrules.RefreshTokenForSharingDuration),
		),
		provider.WithTokenHandler(cdsrules.RecordArrangement),
		provider.WithIDTokenClaims(cdsrules.IDTokenClaims),
		provider.WithTokenIntrospection(cdsrules.IntrospectingClient),
		provider.WithTokenRevocation(allowRevocation),
		provider.WithAuthSessionLifetime(dataHolderFlowLifetimeSeconds),
		provider.WithErrorRenderer(refusedPage(cfg.Pages)),
		provider.WithTokenOptions(func(_ context.Context, grant *goidc.Grant, _ *goidc.Client) goidc.TokenOptions {
			lifetime := cdsrules.AccessTokenLifetime(grant, dataHolderTokenLifetimeSeconds, d.now())

			return goidc.NewJWTTokenOptions(goidc.SigAlgPS256, lifetime)
		}),
		// The library finds clients that are not static only through its DCR
		// manager. The data holder serves the registration endpoints itself.
		provider.WithDCR(registeredClients{cfg.Clients}),
		// The library does not call this handler at the exchange of a code, so flows.GrantByAuthCode checks it.
		provider.WithGrantHandler(func(ctx context.Context, grantType goidc.GrantType, grant *goidc.Grant) error {
			if grantType == goidc.GrantRefreshToken {
				if err := d.requireActive(ctx, grant.ClientID); err != nil {
					return err
				}
			}

			return registrationScopeOnly(ctx, grantType, grant)
		}),
		provider.WithJWTLifetime(dataHolderAssertionLifetimeSeconds),
		provider.WithJWTLeewayTime(clockLeewaySeconds),
		provider.WithJTIConsumer(consumeJTI(cfg.JTIs)),
		provider.WithHTTPClientFunc(func(context.Context) *http.Client { return jwksClient }),
		provider.WithErrorHandler(recordRefusal),
	)
	if err != nil {
		return nil, err
	}

	d.provider = op.Handler()

	return d, nil
}

// RegisterSecure needs a listener that verifies the client certificate.
func (d *DataHolder) RegisterSecure(mux *http.ServeMux) {
	token := cdsrules.TokenResponse(dataHolderTokenPath)(d.provider)
	mux.HandleFunc("POST "+dataHolderTokenPath, tokenHandler(defaultRegistrationScope(token)))
	mux.HandleFunc(
		"POST "+dataHolderIntrospectionPath,
		tokenHandler(cdsrules.IntrospectionResponse(dataHolderIntrospectionPath, d.grants)(d.provider)),
	)
	mux.HandleFunc("POST "+dataHolderRevocationPath, tokenHandler(d.revocation(d.provider)))
	mux.HandleFunc("POST "+dataHolderArrangementRevocationPath, d.arrangementRevocation)
	mux.HandleFunc(
		"POST "+dataHolderPARPath,
		pushedRequestHandler(cdsrules.RequestObjectClaims(dataHolderPARPath)(d.provider)),
	)
}

// RegisterAuthorisation adds the authorisation endpoint and the pages of its flow. The listener must not permit
// cross-origin requests to these routes.
func (d *DataHolder) RegisterAuthorisation(mux *http.ServeMux) {
	mux.Handle("GET "+dataHolderAuthorisationPath, d.provider)
	mux.Handle("GET "+dataHolderAuthorisationPath+"/{flow}", d.provider)
	mux.Handle("POST "+dataHolderAuthorisationPath+"/{flow}", d.provider)
}

// defaultRegistrationScope gives a client credentials request with no scope the only scope of that grant, because the
// library has no default scope.
func defaultRegistrationScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The library reads the form again, and gets the changed values.
		if err := r.ParseForm(); err == nil &&
			r.PostForm.Get("grant_type") == string(goidc.GrantClientCredentials) && r.PostForm.Get("scope") == "" {
			r.PostForm.Set("scope", RegistrationScope)
			r.Form.Set("scope", RegistrationScope)
		}

		next.ServeHTTP(w, r)
	})
}

// The library gives 401 for unauthorized_client, but RFC 6749 gives 400 for it in a token error response.
func (d *DataHolder) requireActive(ctx context.Context, clientID string) error {
	active, err := d.cfg.ClientStatuses.Active(ctx, clientID)
	if err != nil {
		return err
	}

	if !active {
		return goidc.NewError(goidc.ErrorCodeUnauthorizedClient, "the software product is not ACTIVE on the Register").
			WithStatusCode(http.StatusBadRequest)
	}

	return nil
}

// registrationScopeOnly keeps the scopes of the data APIs out of a client credentials token: that grant has no
// consumer.
func registrationScopeOnly(_ context.Context, grantType goidc.GrantType, grant *goidc.Grant) error {
	if grantType == goidc.GrantClientCredentials && grant.Scopes != RegistrationScope {
		return goidc.NewError(
			goidc.ErrorCodeInvalidScope,
			"a client credentials token has only the scope "+RegistrationScope,
		)
	}

	return nil
}

// VerifyAccessToken accepts a token only with the client certificate that the token is bound to (RFC 8705), and only
// when its client is still registered.
func (d *DataHolder) VerifyAccessToken(
	ctx context.Context,
	token string,
	clientCert *x509.Certificate,
) (AccessToken, error) {
	claims, err := d.verifyAccessToken(ctx, token, clientCert)
	if err != nil {
		return AccessToken{}, err
	}

	return claims.accessToken(), nil
}

type ConsentToken struct {
	ClientID string
	Scopes   []string
	// Subject is the sub of the ID token of the consent.
	Subject       string
	ArrangementID string
}

// VerifyConsentToken also refuses a token whose grant does not exist or is revoked.
func (d *DataHolder) VerifyConsentToken(
	ctx context.Context,
	token string,
	clientCert *x509.Certificate,
) (ConsentToken, error) {
	claims, err := d.verifyAccessToken(ctx, token, clientCert)
	if err != nil {
		return ConsentToken{}, err
	}

	// The library does not keep a client credentials grant, so a client credentials token has no grant.
	grant, err := d.cfg.Grants.Grant(ctx, claims.GrantID)
	if errors.Is(err, ErrGrantNotFound) {
		return ConsentToken{}, fmt.Errorf("%w: it is not a token of a consent", ErrInvalidToken)
	}

	if err != nil {
		return ConsentToken{}, err
	}

	if !grant.RevokedAt.IsZero() {
		return ConsentToken{}, fmt.Errorf("%w: its grant is revoked", ErrInvalidToken)
	}

	if grant.ClientID != claims.ClientID {
		return ConsentToken{}, fmt.Errorf("%w: its grant is of a different client", ErrInvalidToken)
	}

	return ConsentToken{
		ClientID:      claims.ClientID,
		Scopes:        strings.Fields(claims.Scope),
		Subject:       grant.Subject,
		ArrangementID: grant.ArrangementID,
	}, nil
}

func (d *DataHolder) verifyAccessToken(
	ctx context.Context,
	token string,
	clientCert *x509.Certificate,
) (accessTokenClaims, error) {
	claims, err := verifyAccessToken(ctx, token, clientCert, d.cfg.Issuer, d.cfg.SigningKeys, d.now())
	if err != nil {
		return accessTokenClaims{}, err
	}

	revoked, err := d.cfg.RevokedTokens.TokenRevoked(ctx, claims.ID)
	if err != nil {
		return accessTokenClaims{}, err
	}

	if revoked {
		return accessTokenClaims{}, fmt.Errorf("%w: it is revoked", ErrInvalidToken)
	}

	_, err = d.cfg.Clients.Client(ctx, claims.ClientID)
	if errors.Is(err, ErrClientNotFound) {
		return accessTokenClaims{}, fmt.Errorf("%w: its client is not registered", ErrInvalidToken)
	}

	if err != nil {
		return accessTokenClaims{}, err
	}

	return claims, nil
}

type registeredClients struct{ store DataHolderClientStore }

func (r registeredClients) Client(ctx context.Context, id string) (*goidc.Client, error) {
	c, err := r.store.Client(ctx, id)
	if errors.Is(err, ErrClientNotFound) {
		return nil, goidc.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	grantTypes := make([]goidc.GrantType, 0, len(c.GrantTypes))
	for _, g := range c.GrantTypes {
		grantTypes = append(grantTypes, goidc.GrantType(g))
	}

	return &goidc.Client{
		ID:               c.ID,
		JWKSURI:          c.JWKSURI,
		ScopeIDs:         strings.Join(c.Scopes, " "),
		GrantTypes:       grantTypes,
		RedirectURIs:     c.RedirectURIs,
		ResponseTypes:    []goidc.ResponseType{goidc.ResponseTypeCode},
		JARSigAlg:        goidc.SignatureAlgorithm(c.RequestObjectSigningAlg),
		TokenAuthnMethod: goidc.AuthnMethodPrivateKeyJWT,
		// The library reads a separate algorithm for each endpoint. The CDR arrangement revocation endpoint
		// authenticates the client at the token revocation endpoint.
		TokenAuthnSigAlg:              goidc.SignatureAlgorithm(c.TokenEndpointAuthSigningAlg),
		TokenIntrospectionAuthnSigAlg: goidc.SignatureAlgorithm(c.TokenEndpointAuthSigningAlg),
		TokenRevocationAuthnSigAlg:    goidc.SignatureAlgorithm(c.TokenEndpointAuthSigningAlg),
	}, nil
}

func (registeredClients) SaveClient(context.Context, *goidc.Client) error {
	return errors.New("authserver: the library does not register clients")
}

func (registeredClients) DeleteClient(context.Context, string) error {
	return errors.New("authserver: the library does not register clients")
}

func (d *DataHolder) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("GET "+dataHolderConfigPath, d.configuration)
	mux.HandleFunc("GET "+dataHolderJWKSPath, jwksHandler(d.cfg.SigningKeys))
}

// AuthorisationPath is the path of the authorisation endpoint. The pages of a flow are below it.
func (d *DataHolder) AuthorisationPath() string {
	return dataHolderAuthorisationPath
}

// UserInfoPath is the path of the UserInfo endpoint on the mTLS listener. The caller serves the endpoint, because the
// claims are data of the customer.
func (d *DataHolder) UserInfoPath() string {
	return dataHolderUserInfoPath
}

// RegistrationPath is the path of the registration endpoint on the mTLS listener.
func (d *DataHolder) RegistrationPath() string {
	return dataHolderRegistrationPath
}

type dataHolderConfigurationJSON struct {
	Issuer                        string   `json:"issuer"`
	JWKSURI                       string   `json:"jwks_uri"`
	AuthorisationEndpoint         string   `json:"authorization_endpoint"`
	RegistrationEndpoint          string   `json:"registration_endpoint"`
	TokenEndpoint                 string   `json:"token_endpoint"`
	UserInfoEndpoint              string   `json:"userinfo_endpoint"`
	IntrospectionEndpoint         string   `json:"introspection_endpoint"`
	RevocationEndpoint            string   `json:"revocation_endpoint"`
	PAREndpoint                   string   `json:"pushed_authorization_request_endpoint"`
	ArrangementRevocationEndpoint string   `json:"cdr_arrangement_revocation_endpoint"`
	Scopes                        []string `json:"scopes_supported"`
	Claims                        []string `json:"claims_supported"`
	ClaimsParameterSupported      bool     `json:"claims_parameter_supported"`
	ACRValues                     []string `json:"acr_values_supported"`
	SubjectTypes                  []string `json:"subject_types_supported"`
	GrantTypes                    []string `json:"grant_types_supported"`
	ResponseTypes                 []string `json:"response_types_supported"`
	ResponseModes                 []string `json:"response_modes_supported"`
	CodeChallengeMethods          []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethods      []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgs  []string `json:"token_endpoint_auth_signing_alg_values_supported"`
	RequestObjectSigningAlgs      []string `json:"request_object_signing_alg_values_supported"`
	IDTokenSigningAlgs            []string `json:"id_token_signing_alg_values_supported"`
	AuthorisationSigningAlgs      []string `json:"authorization_signing_alg_values_supported"`
	CertificateBoundTokens        bool     `json:"tls_client_certificate_bound_access_tokens"`
	RequirePAR                    bool     `json:"require_pushed_authorization_requests"`
}

// configuration writes the document here, because the library omits fields that the CDS requires.
func (d *DataHolder) configuration(w http.ResponseWriter, _ *http.Request) {
	public, secure := d.cfg.Issuer, d.cfg.SecureBaseURL
	clientAlgs := []string{string(goidc.SigAlgPS256), string(goidc.SigAlgES256)}
	ownAlgs := []string{string(goidc.SigAlgPS256)}

	writeJSON(w, dataHolderConfigurationJSON{
		Issuer:                        d.cfg.Issuer,
		JWKSURI:                       public + dataHolderJWKSPath,
		AuthorisationEndpoint:         public + dataHolderAuthorisationPath,
		RegistrationEndpoint:          secure + dataHolderRegistrationPath,
		TokenEndpoint:                 secure + dataHolderTokenPath,
		UserInfoEndpoint:              secure + dataHolderUserInfoPath,
		IntrospectionEndpoint:         secure + dataHolderIntrospectionPath,
		RevocationEndpoint:            secure + dataHolderRevocationPath,
		PAREndpoint:                   secure + dataHolderPARPath,
		ArrangementRevocationEndpoint: secure + dataHolderArrangementRevocationPath,
		Scopes:                        d.cfg.Scopes,
		Claims: []string{
			"sub",
			"acr",
			"auth_time",
			"name",
			"given_name",
			"family_name",
			"updated_at",
			// The CDS calls the sharing duration of a request object a claim, and the request object has it in claims.
			"sharing_duration",
		},
		ClaimsParameterSupported: true,
		ACRValues:                []string{"urn:cds.au:cdr:2"},
		SubjectTypes:             []string{string(goidc.SubIdentifierPairwise)},
		GrantTypes: []string{
			string(goidc.GrantAuthorizationCode),
			string(goidc.GrantClientCredentials),
			string(goidc.GrantRefreshToken),
		},
		ResponseTypes:                []string{string(goidc.ResponseTypeCode)},
		ResponseModes:                []string{string(goidc.ResponseModeJWT)},
		CodeChallengeMethods:         []string{string(goidc.CodeChallengeMethodSHA256)},
		TokenEndpointAuthMethods:     []string{string(goidc.AuthnMethodPrivateKeyJWT)},
		TokenEndpointAuthSigningAlgs: clientAlgs,
		RequestObjectSigningAlgs:     clientAlgs,
		IDTokenSigningAlgs:           ownAlgs,
		AuthorisationSigningAlgs:     ownAlgs,
		CertificateBoundTokens:       true,
		RequirePAR:                   true,
	})
}
