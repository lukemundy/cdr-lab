package certissue_test

import (
	"crypto"
	"crypto/x509"
	"maps"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certissue"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube/kubetest"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register/registertest"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

type lab struct {
	api      *kubetest.API
	register *registertest.Register
	secrets  *kube.Client
	client   *register.Client
}

func newLab(t *testing.T) lab {
	t.Helper()

	api, reg := kubetest.New(t), registertest.New(t)

	return lab{
		api:      api,
		register: reg,
		secrets:  kube.New(api.URL, kubetest.Namespace, api.TokenFile, api.Client()),
		client:   register.New(reg.URL, registertest.Token, reg.Client()),
	}
}

func (l lab) run(t *testing.T, cfg certissue.Config) (bool, error) {
	t.Helper()

	cert, err := certissue.Run(t.Context(), l.secrets, l.client, cfg)

	return cert != nil, err
}

func serverConfig() certissue.Config {
	//nolint:gosec // G101 reads the name of the Kubernetes Secret as a credential
	return certissue.Config{
		Kind:          ca.Server,
		LegalEntityID: "le-1",
		RecordID:      "brand-1",
		CommonName:    "wattle-bank",
		Organisation:  "Wattle Bank Ltd",
		DNSNames:      []string{"wattle-bank", "localhost"},
		Secret:        "wattle-bank-secure-tls",
		RenewBefore:   certissue.DefaultRenewBefore,
		Now:           time.Now,
	}
}

func clientConfig() certissue.Config {
	//nolint:gosec // G101 reads the name of the Kubernetes Secret as a credential
	return certissue.Config{
		Kind:          ca.Client,
		LegalEntityID: "le-2",
		RecordID:      "sp-2",
		CommonName:    "Token Test Active",
		Organisation:  "Token Test",
		Secret:        "conformance-client",
		RenewBefore:   certissue.DefaultRenewBefore,
		Now:           time.Now,
	}
}

func leaf(t *testing.T, s kubetest.Secret) *x509.Certificate {
	t.Helper()

	chain, err := ca.ParseCertificates(s.Data["tls.crt"])
	require.NoError(t, err)

	return chain[0]
}

// Verifies: CLI-0030, CLI-0034
func TestTheCommandWritesTheServerCertificateTheKeyAndTheBundleToATLSSecret(t *testing.T) {
	l := newLab(t)

	issued, err := l.run(t, serverConfig())
	require.NoError(t, err)
	assert.True(t, issued)
	assert.Equal(t, []string{"POST /v1/data-holders/le-1/brands/brand-1/certificates/issue"}, l.register.Issues())

	s, ok := l.api.Secret("wattle-bank-secure-tls")
	require.True(t, ok)
	assert.Equal(t, kube.TLSType, s.Type)
	assert.ElementsMatch(t, []string{"tls.crt", "tls.key", "ca.crt"}, slices.Collect(maps.Keys(s.Data)))

	chain, err := ca.ParseCertificates(s.Data["tls.crt"])
	require.NoError(t, err)
	require.Len(t, chain, 2)
	assert.Equal(t, l.register.Intermediate.Certificate.Raw, chain[1].Raw)
	assert.Equal(t, "CN=wattle-bank,OU=Consumer Data Right,O=Wattle Bank Ltd,C=AU", chain[0].Subject.String())
	assert.ElementsMatch(t, []string{"wattle-bank", "localhost"}, chain[0].DNSNames)

	key, err := ca.ParseKey(s.Data["tls.key"])
	require.NoError(t, err)
	assert.True(t, sameKey(chain[0].PublicKey, key.Public()))

	assert.Equal(
		t,
		ca.EncodeCertificates(l.register.Intermediate.Certificate, l.register.Root.Certificate),
		s.Data["ca.crt"],
	)
}

// Verifies: CLI-0030
func TestTheCommandGetsAClientCertificateForASoftwareProduct(t *testing.T) {
	l := newLab(t)

	_, err := l.run(t, clientConfig())
	require.NoError(t, err)

	assert.Equal(
		t,
		[]string{"POST /v1/data-recipients/le-2/software-products/sp-2/certificates/issue"},
		l.register.Issues(),
	)

	s, _ := l.api.Secret("conformance-client")
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, leaf(t, s).ExtKeyUsage)
}

// Verifies: CLI-0032, CLI-0033
func TestTheCommandSendsOnlyACSRWithANewRSA2048Key(t *testing.T) {
	l := newLab(t)

	_, err := l.run(t, serverConfig())
	require.NoError(t, err)

	csr, err := ca.ParseCSR(l.register.LastCSR())
	require.NoError(t, err)
	require.NoError(t, ca.ValidateCSR(csr, ca.Server))
	assert.NotContains(t, string(l.register.LastCSR()), "PRIVATE KEY")

	s, _ := l.api.Secret("wattle-bank-secure-tls")
	key, err := ca.ParseKey(s.Data["tls.key"])
	require.NoError(t, err)
	assert.True(t, sameKey(csr.PublicKey, key.Public()))
}

// Verifies: CLI-0035, CLI-0062
func TestTheCommandChangesNothingWhenTheCertificateIsCurrent(t *testing.T) {
	for name, cfg := range map[string]certissue.Config{"server": serverConfig(), "client": clientConfig()} {
		t.Run(name, func(t *testing.T) {
			l := newLab(t)
			_, err := l.run(t, cfg)
			require.NoError(t, err)

			before, _ := l.api.Secret(cfg.Secret)

			issued, err := l.run(t, cfg)

			require.NoError(t, err)
			assert.False(t, issued)
			assert.Len(t, l.register.Issues(), 1)

			after, _ := l.api.Secret(cfg.Secret)
			assert.Equal(t, before.ResourceVersion, after.ResourceVersion)
		})
	}
}

// Verifies: CLI-0035
func TestTheCommandRenewsACertificateThatNoLongerMeetsAConditionOfACurrentCertificate(t *testing.T) {
	for name, change := range map[string]func(*testing.T, lab, *certissue.Config){
		"it expires within the renewal window": func(_ *testing.T, _ lab, cfg *certissue.Config) {
			cfg.Now = func() time.Time { return time.Now().AddDate(0, 12, 10) }
		},
		"it does not chain to the CDR CA bundle": func(t *testing.T, l lab, _ *certissue.Config) {
			l.register.NewCA(t)
		},
		"the CN changed": func(_ *testing.T, _ lab, cfg *certissue.Config) { cfg.CommonName = "wattle" },
		"the O changed":  func(_ *testing.T, _ lab, cfg *certissue.Config) { cfg.Organisation = "Wattle Bank" },
		"a DNS name was added": func(_ *testing.T, _ lab, cfg *certissue.Config) {
			cfg.DNSNames = append(cfg.DNSNames, "wattle-bank.cdr-lab.svc")
		},
		"a DNS name was removed": func(_ *testing.T, _ lab, cfg *certissue.Config) { cfg.DNSNames = cfg.DNSNames[:1] },
		"the key does not agree": func(_ *testing.T, l lab, cfg *certissue.Config) {
			s, _ := l.api.Secret(cfg.Secret)
			other, _ := l.api.Secret("other")
			s.Data["tls.key"] = other.Data["tls.key"]
			l.api.Put(cfg.Secret, s)
		},
	} {
		t.Run(name, func(t *testing.T) {
			l := newLab(t)
			cfg := serverConfig()

			other := cfg
			other.Secret = "other"
			_, err := l.run(t, other)
			require.NoError(t, err)

			_, err = l.run(t, cfg)
			require.NoError(t, err)

			change(t, l, &cfg)

			issued, err := l.run(t, cfg)

			require.NoError(t, err)
			assert.True(t, issued)

			s, _ := l.api.Secret(cfg.Secret)
			assert.Equal(t, cfg.CommonName, leaf(t, s).Subject.CommonName)
		})
	}
}

// Verifies: CLI-0036
func TestARenewalWindowFromTheInputsReplacesTheDefault(t *testing.T) {
	l := newLab(t)
	cfg := serverConfig()
	_, err := l.run(t, cfg)
	require.NoError(t, err)

	cfg.RenewBefore = 400 * 24 * time.Hour

	issued, err := l.run(t, cfg)

	require.NoError(t, err)
	assert.True(t, issued)
}

func TestTheDefaultRenewalWindowIsThirtyDays(t *testing.T) {
	assert.Equal(t, 30*24*time.Hour, certissue.DefaultRenewBefore)
}

// Verifies: CLI-0037, CLI-0038
func TestAFailedIssueChangesNoSecretAndGivesTheStatusAndTheDetail(t *testing.T) {
	l := newLab(t)
	cfg := serverConfig()
	_, err := l.run(t, cfg)
	require.NoError(t, err)

	before, _ := l.api.Secret(cfg.Secret)
	cfg.CommonName = "wattle"

	l.register.Refuse(http.StatusBadRequest, "The O must be the brand name or the legal entity name.")

	_, err = l.run(t, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "400 Bad Request")
	assert.Contains(t, err.Error(), "The O must be the brand name or the legal entity name.")

	after, _ := l.api.Secret(cfg.Secret)
	assert.Equal(t, before.ResourceVersion, after.ResourceVersion)
}

// Verifies: CLI-0037
func TestARegisterThatIsNotReadyChangesNoSecret(t *testing.T) {
	l := newLab(t)
	l.register.Close()

	_, err := l.run(t, serverConfig())

	require.Error(t, err)
	assert.Empty(t, l.api.Names())
}

// Verifies: CLI-0039
func TestARenewalDoesNotRevokeTheReplacedCertificate(t *testing.T) {
	l := newLab(t)
	cfg := serverConfig()
	_, err := l.run(t, cfg)
	require.NoError(t, err)

	cfg.CommonName = "wattle"
	_, err = l.run(t, cfg)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"POST /v1/data-holders/le-1/brands/brand-1/certificates/issue",
		"POST /v1/data-holders/le-1/brands/brand-1/certificates/issue",
	}, l.register.Issues())
}

func TestASecretOfADifferentTypeIsAnErrorBeforeTheIssue(t *testing.T) {
	l := newLab(t)
	cfg := serverConfig()
	l.api.Put(cfg.Secret, kubetest.Secret{Type: "Opaque", Data: map[string][]byte{"x": []byte("y")}})

	_, err := l.run(t, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubernetes.io/tls")
	assert.Empty(t, l.register.Issues())
}

func TestAFailedWriteGivesTheErrorOfTheKubernetesAPI(t *testing.T) {
	l := newLab(t)
	cfg := serverConfig()
	l.api.Refuse(http.MethodPost, cfg.Secret, http.StatusForbidden)

	_, err := l.run(t, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func sameKey(a, b crypto.PublicKey) bool {
	k, ok := a.(interface{ Equal(x crypto.PublicKey) bool })
	return ok && k.Equal(b)
}
