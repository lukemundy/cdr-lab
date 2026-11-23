package cainit_test

import (
	"crypto"
	"maps"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/cainit"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube/kubetest"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

var now = time.Date(2026, 11, 23, 4, 0, 0, 0, time.UTC)

func setup(t *testing.T) (*kubetest.API, func() (bool, error)) {
	t.Helper()

	api := kubetest.New(t)
	secrets := kube.New(api.URL, kubetest.Namespace, api.TokenFile, api.Client())
	//nolint:gosec // G101 reads the names of the Kubernetes Secrets as credentials
	cfg := cainit.Config{
		IntermediateSecret: "cdr-intermediate",
		CABundleSecret:     "cdr-ca-bundle",
		KeyBits:            2048,
		Now:                func() time.Time { return now },
	}

	return api, func() (bool, error) { return cainit.Run(t.Context(), secrets, cfg) }
}

// Verifies: CLI-0010, CLI-0011, CLI-0012, CLI-0013
func TestTheCASetUpWritesTheIntermediateAndTheBundleSecrets(t *testing.T) {
	api, run := setup(t)

	created, err := run()
	require.NoError(t, err)
	assert.True(t, created)

	intermediate, ok := api.Secret("cdr-intermediate")
	require.True(t, ok)
	assert.Equal(t, kube.TLSType, intermediate.Type)
	assert.ElementsMatch(t, []string{"tls.crt", "tls.key"}, keys(intermediate.Data))

	certs, err := ca.ParseCertificates(intermediate.Data["tls.crt"])
	require.NoError(t, err)
	require.Len(t, certs, 1)
	assert.Equal(t, "CN=Mock CDR Intermediate CA,OU=Consumer Data Right,O=cdr-lab,C=AU", certs[0].Subject.String())
	assert.Equal(t, now.AddDate(10, 0, 0), certs[0].NotAfter)

	key, err := ca.ParseKey(intermediate.Data["tls.key"])
	require.NoError(t, err)
	assert.True(t, sameKey(certs[0].PublicKey, key.Public()))

	bundle, ok := api.Secret("cdr-ca-bundle")
	require.True(t, ok)
	assert.Equal(t, []string{"ca.crt"}, keys(bundle.Data))

	chain, err := ca.ParseCertificates(bundle.Data["ca.crt"])
	require.NoError(t, err)
	require.Len(t, chain, 2)
	assert.Equal(t, certs[0].Raw, chain[0].Raw)
	assert.Equal(t, "CN=Mock CDR Root CA,OU=Consumer Data Right,O=cdr-lab,C=AU", chain[1].Subject.String())
	require.NoError(t, chain[0].CheckSignatureFrom(chain[1]))
}

// Verifies: CLI-0016
func TestNoSecretHasTheRootKey(t *testing.T) {
	api, run := setup(t)

	_, err := run()
	require.NoError(t, err)

	assert.ElementsMatch(t, []string{"cdr-intermediate", "cdr-ca-bundle"}, api.Names())

	bundle, _ := api.Secret("cdr-ca-bundle")
	chain, err := ca.ParseCertificates(bundle.Data["ca.crt"])
	require.NoError(t, err)

	root := chain[1]

	for _, name := range api.Names() {
		s, _ := api.Secret(name)
		for field, data := range s.Data {
			key, err := ca.ParseKey(data)
			if err != nil {
				continue
			}

			assert.False(t, sameKey(root.PublicKey, key.Public()), "%s/%s", name, field)
		}
	}
}

// Verifies: CLI-0014, CLI-0062
func TestTheCASetUpChangesNothingWhenBothSecretsExist(t *testing.T) {
	api, run := setup(t)
	_, err := run()
	require.NoError(t, err)

	before, _ := api.Secret("cdr-intermediate")

	created, err := run()

	require.NoError(t, err)
	assert.False(t, created)

	after, _ := api.Secret("cdr-intermediate")
	assert.Equal(t, before.ResourceVersion, after.ResourceVersion)
}

// Verifies: CLI-0015
func TestTheCASetUpFailsAndNamesTheAbsentSecretWhenOnlyOneExists(t *testing.T) {
	for _, tc := range []struct{ present, absent string }{
		{present: "cdr-intermediate", absent: "cdr-ca-bundle"},
		{present: "cdr-ca-bundle", absent: "cdr-intermediate"},
	} {
		t.Run(tc.absent, func(t *testing.T) {
			api, run := setup(t)
			api.Put(tc.present, kubetest.Secret{Type: "Opaque", Data: map[string][]byte{"x": []byte("y")}})

			_, err := run()

			require.Error(t, err)
			assert.Contains(t, err.Error(), `"`+tc.absent+`"`)
			assert.Equal(t, []string{tc.present}, api.Names())
		})
	}
}

// Verifies: CLI-0017
func TestAFailedWriteOfTheSecondSecretDeletesTheFirst(t *testing.T) {
	api, run := setup(t)
	api.Refuse(http.MethodPost, "cdr-ca-bundle", http.StatusForbidden)

	_, err := run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cdr-ca-bundle")
	assert.Empty(t, api.Names())
}

func TestTheCASetUpGivesTheErrorOfAFailedRead(t *testing.T) {
	api, run := setup(t)
	api.Refuse(http.MethodGet, "cdr-intermediate", http.StatusForbidden)

	_, err := run()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
	assert.Empty(t, api.Names())
}

func keys(data map[string][]byte) []string {
	return slices.Collect(maps.Keys(data))
}

func sameKey(a, b crypto.PublicKey) bool {
	k, ok := a.(interface{ Equal(x crypto.PublicKey) bool })
	return ok && k.Equal(b)
}
