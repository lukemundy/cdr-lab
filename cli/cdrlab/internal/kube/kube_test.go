package kube_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube/kubetest"
)

func client(t *testing.T) (*kube.Client, *kubetest.API) {
	t.Helper()

	api := kubetest.New(t)

	return kube.New(api.URL, kubetest.Namespace, api.TokenFile, api.Client()), api
}

func TestGetGivesASecretThatExists(t *testing.T) {
	c, api := client(t)
	api.Put("tls", kubetest.Secret{Type: kube.TLSType, Data: map[string][]byte{"tls.crt": []byte("cert")}})

	s, found, err := c.Get(t.Context(), "tls")

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "tls", s.Name)
	assert.Equal(t, kube.TLSType, s.Type)
	assert.Equal(t, []byte("cert"), s.Data["tls.crt"])
	assert.NotEmpty(t, s.ResourceVersion)
}

func TestGetOfAnAbsentSecretIsNotAnError(t *testing.T) {
	c, _ := client(t)

	_, found, err := c.Get(t.Context(), "absent")

	require.NoError(t, err)
	assert.False(t, found)
}

func TestCreateUpdateAndDeleteChangeTheSecret(t *testing.T) {
	c, api := client(t)

	require.NoError(t, c.Create(t.Context(), kube.Secret{
		Name: "tls", Type: kube.TLSType, Data: map[string][]byte{"tls.key": []byte("one")},
	}))

	s, _, err := c.Get(t.Context(), "tls")
	require.NoError(t, err)

	s.Data = map[string][]byte{"tls.key": []byte("two")}
	require.NoError(t, c.Update(t.Context(), s))

	stored, _ := api.Secret("tls")
	assert.Equal(t, []byte("two"), stored.Data["tls.key"])

	require.NoError(t, c.Delete(t.Context(), "tls"))
	assert.Empty(t, api.Names())
}

func TestUpdateRefusesAStaleSecret(t *testing.T) {
	c, api := client(t)
	api.Put("tls", kubetest.Secret{Type: kube.TLSType})

	s, _, err := c.Get(t.Context(), "tls")
	require.NoError(t, err)
	api.Put("tls", kubetest.Secret{Type: kube.TLSType})

	err = c.Update(t.Context(), s)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "409")
}

func TestAnErrorGivesTheVerbTheSecretAndTheMessageOfTheAPI(t *testing.T) {
	c, api := client(t)
	api.Refuse(http.MethodGet, "tls", http.StatusForbidden)

	_, _, err := c.Get(t.Context(), "tls")

	require.Error(t, err)
	assert.Contains(t, err.Error(), `get the Secret "tls"`)
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), `refused GET of secrets "tls"`)
}

// Verifies: CLI-0001
func TestEachRequestReadsTheTokenFileAgain(t *testing.T) {
	c, api := client(t)
	require.NoError(t, os.WriteFile(api.TokenFile, []byte("rotated-token"), 0o600))

	_, _, err := c.Get(t.Context(), "tls")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

// Verifies: CLI-0001, CLI-0002
func TestInClusterUsesTheServiceAccountAndTheNamespaceOfThePod(t *testing.T) {
	api := kubetest.NewTLS(t)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "namespace"), []byte(kubetest.Namespace), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte(kubetest.Token), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.crt"), api.CAPEM(), 0o600))
	t.Setenv("KUBERNETES_SERVICE_HOST", api.Host())
	t.Setenv("KUBERNETES_SERVICE_PORT", api.Port())

	c, err := kube.InCluster(dir)
	require.NoError(t, err)

	_, found, err := c.Get(t.Context(), "tls")
	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, kubetest.Namespace, c.Namespace())
}

func TestInClusterFailsOutsideAPod(t *testing.T) {
	t.Setenv("KUBERNETES_SERVICE_HOST", "")

	_, err := kube.InCluster(t.TempDir())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "KUBERNETES_SERVICE_HOST")
}
