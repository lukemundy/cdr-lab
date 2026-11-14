package app_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
	"github.com/lukemundy/cdr-lab/pkg/eventsourcing"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

var issueTime = time.Date(2026, 10, 27, 6, 30, 15, 500, time.UTC)

// RSA key generation is slow, so the tests share one lab CA and one CSR key.
var testLab = sync.OnceValue(func() ca.Lab {
	var keys [3]*rsa.PrivateKey

	for i := range keys {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}

		keys[i] = k
	}

	lab, err := ca.NewLab(ca.LabKeys{CDRRoot: keys[0], CDRIntermediate: keys[1], PublicRoot: keys[2]}, issueTime)
	if err != nil {
		panic(err)
	}

	lab.CDRIntermediate.CRLURL = "http://ca.example/crl"

	return lab
})

var csrKey = sync.OnceValue(func() *rsa.PrivateKey {
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	return k
})

// clientCSR has the names of sp-1 of seededWithProduct.
func clientCSR(t *testing.T, change func(*pkix.Name)) *x509.CertificateRequest {
	t.Helper()

	subject := pkix.Name{
		CommonName:         "App",
		Organization:       []string{"App"},
		OrganizationalUnit: []string{"Consumer Data Right"},
		Country:            []string{"AU"},
	}
	if change != nil {
		change(&subject)
	}

	der, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{Subject: subject}, csrKey())
	require.NoError(t, err)
	csr, err := x509.ParseCertificateRequest(der)
	require.NoError(t, err)

	return csr
}

func certificateIssuer(store *memoryStore) app.CertificateIssuer {
	return app.NewCertificateIssuer(store, testLab().CDRIntermediate, func() time.Time { return issueTime })
}

func TestTheIssueGivesACertificateForTheCSRAndRegistersIt(t *testing.T) {
	store := seededWithProduct(t)

	chain, err := certificateIssuer(store).IssueClient(t.Context(), "dr-1", "sp-1", clientCSR(t, nil))

	require.NoError(t, err)
	require.Len(t, chain, 2)
	cert := chain[0]
	assert.Equal(t, testLab().CDRIntermediate.Certificate.Raw, chain[1].Raw, "the intermediate follows the certificate")
	assert.Equal(t, "App", cert.Subject.CommonName)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, cert.ExtKeyUsage)
	assert.Equal(t, []string{"http://ca.example/crl"}, cert.CRLDistributionPoints)
	assert.Equal(t, issueTime.Truncate(time.Second), cert.NotBefore)
	require.NoError(t, cert.CheckSignatureFrom(testLab().CDRIntermediate.Certificate))

	sum := sha256.Sum256(cert.Raw)
	last := store.appends[len(store.appends)-1]
	assert.Equal(t, "dr-1", last.StreamID)
	assert.Equal(t, int64(1), last.ExpectedVersion)
	require.Len(t, last.Events, 1)
	assert.Equal(t, "ClientCertificateIssued", last.Events[0].EventType)

	state, _, err := app.DataRecipientType.Load(t.Context(), store, "dr-1")
	require.NoError(t, err)
	assert.Equal(
		t,
		[]string{base64.RawURLEncoding.EncodeToString(sum[:])},
		state.DataRecipient.Brands[0].SoftwareProducts[0].ClientCertificates,
	)
}

func TestTheIssueNamesEachRuleOfTheProfileAndOfTheRecordThatTheCSRBreaks(t *testing.T) {
	store := seededWithProduct(t)
	appends := len(store.appends)

	_, err := certificateIssuer(store).IssueClient(t.Context(), "dr-1", "sp-1", clientCSR(t, func(n *pkix.Name) {
		n.CommonName = "Other"
		n.OrganizationalUnit = nil
	}))

	var rules domain.CSRRulesError
	require.ErrorAs(t, err, &rules)
	assert.Equal(t, []string{
		`subject OU must be "Consumer Data Right"`,
		`subject CN must be the software product name "App"`,
	}, rules.Rules)
	assert.Len(t, store.appends, appends, "a refused issue changes nothing")
}

func TestTheIssueIsRefusedForARemovedSoftwareProductBeforeTheCSRRules(t *testing.T) {
	store := seededWithProduct(t)
	require.NoError(t, app.NewOperator(store).Execute(
		t.Context(),
		"dr-1",
		domain.RemoveSoftwareProduct{SoftwareProductID: "sp-1"},
	))
	appends := len(store.appends)

	_, err := certificateIssuer(
		store,
	).IssueClient(t.Context(), "dr-1", "sp-1", clientCSR(t, func(n *pkix.Name) { n.Country = nil }))

	require.ErrorAs(t, err, new(domain.StatusRuleError))
	assert.Len(t, store.appends, appends)
}

func TestTheIssueIsRefusedForASoftwareProductThatDoesNotExist(t *testing.T) {
	store := seededWithProduct(t)

	for _, ids := range [][2]string{{"dr-1", "sp-unknown"}, {"dr-unknown", "sp-1"}} {
		_, err := certificateIssuer(store).IssueClient(t.Context(), ids[0], ids[1], clientCSR(t, nil))
		require.ErrorIs(t, err, domain.ErrNotFound, ids)
	}
}

// seededWithBrands has the data holder dh-1 with dhb-1 (ACTIVE), dhb-inactive and interim-1 (no dataHolderBrandId).
func seededWithBrands(t *testing.T) *memoryStore {
	t.Helper()

	brand := domain.DataHolderBrand{
		DataHolderBrandID: "dhb-1",
		BrandName:         "Holder Bank",
		Industries:        []domain.Industry{domain.Banking},
		LogoURI:           "https://holder.example/brand.png",
		Status:            domain.Active,
		EndpointDetail: domain.EndpointDetail{
			Version:         "v1",
			PublicBaseURI:   "https://api.holder.example",
			ResourceBaseURI: "https://mtls.holder.example",
			InfosecBaseURI:  "https://id.holder.example",
			WebsiteURI:      "https://holder.example",
		},
		JWKSEndpoint: "https://holder.example/jwks",
	}
	inactive, interim := brand, brand
	inactive.DataHolderBrandID, inactive.Status = "dhb-inactive", domain.Inactive
	interim.DataHolderBrandID, interim.InterimID = "", "interim-1"

	p := participants()
	p.DataHolders[0].Brands = []domain.DataHolderBrand{brand, inactive, interim}
	store := &memoryStore{}
	require.NoError(t, app.Seed(t.Context(), store, p))

	return store
}

// serverCSR has the legal entity name of dh-1 as O.
func serverCSR(t *testing.T, change func(*x509.CertificateRequest)) *x509.CertificateRequest {
	t.Helper()

	tmpl := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:         "mtls.holder.example",
			Organization:       []string{"Holder"},
			OrganizationalUnit: []string{"Consumer Data Right"},
			Country:            []string{"AU"},
		},
		DNSNames: []string{"localhost"},
	}
	if change != nil {
		change(tmpl)
	}

	der, err := x509.CreateCertificateRequest(rand.Reader, tmpl, csrKey())
	require.NoError(t, err)
	csr, err := x509.ParseCertificateRequest(der)
	require.NoError(t, err)

	return csr
}

func TestTheServerIssueGivesACertificateForTheCSRAndRecordsItForTheBrand(t *testing.T) {
	store := seededWithBrands(t)

	chain, err := certificateIssuer(store).IssueServer(t.Context(), "dh-1", "dhb-1", serverCSR(t, nil))

	require.NoError(t, err)
	require.Len(t, chain, 2)
	cert := chain[0]
	assert.Equal(t, testLab().CDRIntermediate.Certificate.Raw, chain[1].Raw, "the intermediate follows the certificate")
	assert.Equal(t, "mtls.holder.example", cert.Subject.CommonName)
	assert.Equal(t, []string{"mtls.holder.example", "localhost"}, cert.DNSNames)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage)
	assert.Equal(t, []string{"http://ca.example/crl"}, cert.CRLDistributionPoints)
	assert.Equal(t, issueTime.Truncate(time.Second), cert.NotBefore)
	require.NoError(t, cert.CheckSignatureFrom(testLab().CDRIntermediate.Certificate))

	last := store.appends[len(store.appends)-1]
	assert.Equal(t, "data-holder", last.StreamType)
	assert.Equal(t, "dh-1", last.StreamID)
	assert.Equal(t, int64(1), last.ExpectedVersion)
	require.Len(t, last.Events, 1)
	assert.Equal(t, "ServerCertificateIssued", last.Events[0].EventType)

	recorded, err := store.Read(t.Context(), "data-holder", "dh-1")
	require.NoError(t, err)
	event, err := app.DataHolderType.Codec.Decode(recorded[len(recorded)-1].EventData)
	require.NoError(t, err)
	assert.Equal(
		t,
		domain.ServerCertificateIssued{BrandID: "dhb-1", Thumbprint: domain.CertificateThumbprint(cert)},
		event,
	)
}

func TestTheServerIssueIsPermittedForAnInactiveBrand(t *testing.T) {
	store := seededWithBrands(t)

	_, err := certificateIssuer(
		store,
	).IssueServer(t.Context(), "dh-1", "dhb-inactive", serverCSR(t, func(r *x509.CertificateRequest) {
		r.Subject.Organization = []string{"Holder Bank"}
	}))

	require.NoError(t, err)
}

func TestTheServerIssueNamesEachRuleOfTheProfileAndOfTheRecordThatTheCSRBreaks(t *testing.T) {
	store := seededWithBrands(t)
	appends := len(store.appends)

	_, err := certificateIssuer(
		store,
	).IssueServer(t.Context(), "dh-1", "dhb-1", serverCSR(t, func(r *x509.CertificateRequest) {
		r.Subject.Organization = []string{"Other"}
		r.DNSNames = []string{"*.holder.example"}
	}))

	var rules domain.CSRRulesError
	require.ErrorAs(t, err, &rules)
	assert.Len(t, rules.Rules, 2)
	assert.Contains(t, rules.Rules[0], `DNS name "*.holder.example" is not a host name`)
	assert.Equal(t, `subject O must be the brand name "Holder Bank" or the legal entity name "Holder"`, rules.Rules[1])
	assert.Len(t, store.appends, appends, "a refused issue changes nothing")
}

func TestTheServerIssueIsRefusedForABrandThatCannotUseItBeforeTheCSRRules(t *testing.T) {
	store := seededWithBrands(t)
	require.NoError(t, app.NewOperator(store).ExecuteDataHolder(
		t.Context(),
		"dh-1",
		domain.RemoveDataHolderBrand{BrandID: "dhb-inactive"},
	))
	appends := len(store.appends)
	csr := serverCSR(t, func(r *x509.CertificateRequest) { r.Subject.Country = nil })

	for _, id := range []string{"dhb-inactive", "interim-1"} {
		_, err := certificateIssuer(store).IssueServer(t.Context(), "dh-1", id, csr)
		require.ErrorAs(t, err, new(domain.StatusRuleError), id)
	}

	for _, ids := range [][2]string{{"dh-1", "dhb-unknown"}, {"dh-unknown", "dhb-1"}} {
		_, err := certificateIssuer(store).IssueServer(t.Context(), ids[0], ids[1], csr)
		require.ErrorIs(t, err, domain.ErrNotFound, ids)
	}

	assert.Len(t, store.appends, appends)
}

func TestTheIssueAtTheVersionThatTheOperatorSawGivesACertificate(t *testing.T) {
	clients := seededWithProduct(t)
	servers := seededWithBrands(t)
	issuer := certificateIssuer(clients)

	client, err := issuer.IssueClientAt(t.Context(), "dr-1", 1, "sp-1", clientCSR(t, nil))
	require.NoError(t, err)
	assert.Len(t, client, 2)

	server, err := certificateIssuer(servers).IssueServerAt(t.Context(), "dh-1", 1, "dhb-1", serverCSR(t, nil))
	require.NoError(t, err)
	assert.Len(t, server, 2)
}

func TestTheIssueAtAStaleVersionGivesNoCertificateAndChangesNothing(t *testing.T) {
	clients := seededWithProduct(t)
	servers := seededWithBrands(t)
	appends := len(clients.appends) + len(servers.appends)

	client, err := certificateIssuer(clients).IssueClientAt(t.Context(), "dr-1", 2, "sp-1", clientCSR(t, nil))
	require.ErrorIs(t, err, eventsourcing.ErrVersionConflict)
	assert.Nil(t, client)

	server, err := certificateIssuer(servers).IssueServerAt(t.Context(), "dh-1", 2, "dhb-1", serverCSR(t, nil))
	require.ErrorIs(t, err, eventsourcing.ErrVersionConflict)
	assert.Nil(t, server)

	assert.Equal(t, appends, len(clients.appends)+len(servers.appends))
}
