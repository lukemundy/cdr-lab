package root

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/cds"
	"github.com/lukemundy/cdr-lab/pkg/cds/revocation"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/telemetry"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/arrangementstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/banksite"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/consumerpages"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/disclosurestore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/httpapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/jtistore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/operatorapi"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/overviewstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/productstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationjwt"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/registrationstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/revokedtokenstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/smtpmail"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/subjectstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/transactionstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func apiListeners(ctx context.Context, pool *pgxpool.Pool, cfg apiConfig, parts shared) ([]httpserver.Listener, error) {
	jtis := jtistore.New(pool)
	transactions := transactionstore.New(pool)
	disclosures := disclosurestore.New(pool)

	mailer, err := smtpmail.New(cfg.smtpAddress, cfg.mailFrom, smtpmail.Brand{
		Name: cfg.brand.Name, Logo: cfg.brand.LogoPNG, Colours: smtpmail.RoleColours(cfg.brand.Colours),
	})
	if err != nil {
		return nil, fmt.Errorf("create the mailer: %w", err)
	}

	clientStatuses := app.ClientStatuses{Registrations: parts.registrations, Statuses: parts.statuses}
	consentData := app.ConsentData{
		Arrangements: parts.events, Customers: parts.customers, TransactionStore: transactions,
		Now: time.Now,
	}

	authServer, err := authserver.NewDataHolder(authserver.DataHolderConfig{
		Issuer:        cfg.publicBaseURL,
		SecureBaseURL: cfg.secureBaseURL,
		Scopes:        domain.SupportedScopes(),
		Clients:       registrationstore.Clients{Reader: parts.registrations},
		JTIs:          jtistore.AuthServer{Store: jtis},
		SigningKeys:   parts.signingKeys.SigningKeys,
		Pages: consumerpages.New(
			cfg.brand.Name,
			parts.registrations,
			app.Authenticator{
				Customers: parts.customers,
				Mail:      mailer,
				Random:    randomUint64,
				FixedOTPs: cfg.fixedOTPs,
				Now:       time.Now,
			},
			app.Consents{
				Customers: parts.customers,
				Subjects:  subjectstore.New(pool),
				Store:     parts.events,
				NewID:     uuid.NewString,
				Now:       time.Now,
			},
			time.Now,
		),
		Grants:         parts.grants,
		RevokedTokens:  revokedtokenstore.New(pool),
		Arrangements:   arrangementRevoker{parts.arrangements},
		ClientStatuses: clientStatuses,
	})
	if err != nil {
		return nil, fmt.Errorf("create the authorisation server: %w", err)
	}

	products := productstore.New(pool)
	api := http.NewServeMux()
	httpapi.Register(api, cfg.publicBaseURL, products, time.Now)
	authServer.RegisterPublic(api)
	authServer.RegisterAuthorisation(api)
	cfg.brand.Register(api)
	consumerpages.RegisterStatic(api, consumerpages.Static())

	registrar := app.Registrar{
		Issuer: cfg.publicBaseURL,
		JWTs: registrationjwt.NewVerifier(
			registrationjwt.NewRegisterKeys(cfg.registerJWKSURL, http.DefaultTransport),
			http.DefaultTransport,
		),
		JTIs:          jtis,
		Store:         parts.events,
		Registrations: parts.registrations,
		Statuses:      parts.statuses,
		Now:           time.Now,
		NewID:         uuid.NewString,
	}

	secure := http.NewServeMux()
	httpapi.RegisterSecure(secure, httpapi.Secure{
		ResourceBaseURL:  cfg.secureBaseURL,
		RegistrationPath: authServer.RegistrationPath(),
		UserInfoPath:     authServer.UserInfoPath(),
		Registrar:        registrar,
		Tokens:           authServer,
		Registrations:    parts.registrations,
		ConsentTokens:    authServer,
		Customers:        consentData,
		Accounts:         consentData,
		Transactions:     consentData,
		Balances:         consentData,
		ClientStatuses:   clientStatuses,
		Disclosures:      app.Disclosures{Store: disclosures, Now: time.Now},
	})
	authServer.RegisterSecure(secure)

	operator := http.NewServeMux()
	customers := app.Customers{
		Store: parts.customers, Arrangements: parts.arrangements, Transactions: transactions, Random: randomUint64,
		NewID: uuid.NewString, Now: time.Now, Products: products, BSB: cfg.brand.BSB,
	}
	operatorapi.Register(
		operator,
		customers,
		parts.arrangements,
		cfg.operator,
	)

	clientCAs := x509.NewCertPool()
	for _, ca := range cfg.cdrCAs {
		clientCAs.AddCert(ca)
	}

	revocations := revocation.NewChecker(cfg.cdrCAs)
	revocations.MaxAge = cfg.crlMaxAge

	consoles, err := consoleListeners(cfg.console, console.UI(), console.Ports{
		BrandName:  cfg.brand.Name,
		BrandFiles: func(routes *httproute.Routes) { cfg.brand.RegisterConsole(routes) },
		Customers:  parts.customers,
		Creates:    customers,
		Details: app.CustomerDetails{
			Customers: parts.customers, Transactions: transactions, Arrangements: parts.arrangements.Reader,
			Registrations: parts.registrations,
		},
		Accounts:           customers,
		Arrangements:       parts.arrangements,
		Registrations:      parts.registrations,
		ClientArrangements: arrangementstore.NewQueries(pool),
		Statuses:           parts.statuses,
		Overview:           overviewstore.New(pool),
		Now:                time.Now,
	})
	if err != nil {
		return nil, err
	}

	customerData := app.CustomerData{
		Customers: parts.customers, TransactionStore: transactions, Arrangements: parts.arrangements.Reader,
		Disclosures: disclosures, Now: time.Now,
	}

	sites, err := bankSiteListeners(cfg.bankSite, cfg.publicCert, banksite.UI(), banksite.Ports{
		BrandName:   cfg.brand.Name,
		BrandLayout: cfg.brand.Layout,
		BrandPhotos: cfg.brand.Photos(),
		BrandFiles:  func(routes *httproute.Routes) { cfg.brand.RegisterSite(routes) },
		Products:    products,
		BrandID:     cfg.brandID,
		Customers:   customers,
		Opener:      customers,
		Accounts:    customerData,
		Consents:    customerData,
		Withdrawer:  parts.arrangements,
		Now:         time.Now,
	})
	if err != nil {
		return nil, err
	}

	listeners := append([]httpserver.Listener{
		{
			Port: apiPort,
			Handler: httpapi.AllowAnyOrigin(
				cds.InteractionID(telemetry.NewHTTPHandler(api)),
				authServer.AuthorisationPath(),
			),
			TLSConfig: cds.TLSConfig(cfg.publicCert),
		},
		{Port: operatorPort, Handler: telemetry.NewHTTPHandler(operator)},
		{
			Port:      securePort,
			Handler:   cds.InteractionID(telemetry.NewHTTPHandler(secure)),
			TLSConfig: cds.MutualTLSConfig(ctx, cfg.secureCert, clientCAs, revocations.Check),
		},
	}, consoles...)

	return append(listeners, sites...), nil
}

// arrangementRevoker gives the authorisation server the error that it knows.
type arrangementRevoker struct{ arrangements app.Arrangements }

func (a arrangementRevoker) RevokeArrangement(ctx context.Context, clientID, arrangementID string) error {
	err := a.arrangements.RevokeByDataRecipient(ctx, clientID, arrangementID)
	if errors.Is(err, app.ErrInvalidArrangement) {
		return authserver.ErrInvalidArrangement
	}

	return err
}

func randomUint64() uint64 {
	var random [8]byte

	// Read gives no error: the Go runtime stops the process when the system has no random source.
	_, _ = rand.Read(random[:])

	return binary.BigEndian.Uint64(random[:])
}
