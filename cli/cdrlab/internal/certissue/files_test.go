package certissue_test

import (
	"crypto/rand"
	"crypto/rsa"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certcheck"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certissue"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/pemfile"
	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func filesIn(t *testing.T) certissue.Files {
	t.Helper()

	dir := t.TempDir()

	return certissue.Files{
		Cert:    filepath.Join(dir, "wattle-bank", "secure.crt"),
		Key:     filepath.Join(dir, "wattle-bank", "secure.key"),
		KeyPerm: pemfile.OwnerOnly,
	}
}

func (l lab) runFiles(t *testing.T, out certissue.Files, cfg certissue.Config) (bool, error) {
	t.Helper()

	cert, err := certissue.Run(t.Context(), out, l.client, cfg)

	return cert != nil, err
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	require.NoError(t, err)

	return data
}

// Verifies: CLI-0030, CLI-0124, CLI-0125
func TestWithFilesTheCommandWritesTheChainAndAKeyThatOnlyTheOwnerCanRead(t *testing.T) {
	l, out := newLab(t), filesIn(t)

	issued, err := l.runFiles(t, out, serverConfig().Config)
	require.NoError(t, err)
	assert.True(t, issued)
	assert.Equal(t, []string{"POST /v1/data-holders/le-1/brands/brand-1/certificates/issue"}, l.register.Issues())

	chain, err := ca.ReadCertificates(out.Cert)
	require.NoError(t, err)
	require.Len(t, chain, 2)
	assert.Equal(t, l.register.Intermediate.Certificate.Raw, chain[1].Raw)

	key, err := ca.ReadKey(out.Key)
	require.NoError(t, err)
	assert.True(t, sameKey(chain[0].PublicKey, key.Public()))

	assert.Equal(t, os.FileMode(0o600), mode(t, out.Key))
	assert.Equal(t, os.FileMode(0o644), mode(t, out.Cert))
}

// Verifies: CLI-0125
func TestWithFilesAnInputLetsAnyUserReadTheKeyFile(t *testing.T) {
	l, out := newLab(t), filesIn(t)
	out.KeyPerm = pemfile.AnyUser

	_, err := l.runFiles(t, out, clientConfig().Config)
	require.NoError(t, err)

	assert.Equal(t, os.FileMode(0o644), mode(t, out.Key))
}

func mode(t *testing.T, path string) os.FileMode {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)

	return info.Mode().Perm()
}

// Verifies: CLI-0035, CLI-0062
func TestWithFilesTheCommandChangesNothingWhenTheCertificateIsCurrent(t *testing.T) {
	l, out := newLab(t), filesIn(t)
	_, err := l.runFiles(t, out, serverConfig().Config)
	require.NoError(t, err)

	cert, key := readFile(t, out.Cert), readFile(t, out.Key)

	issued, err := l.runFiles(t, out, serverConfig().Config)

	require.NoError(t, err)
	assert.False(t, issued)
	assert.Len(t, l.register.Issues(), 1)
	assert.Equal(t, cert, readFile(t, out.Cert))
	assert.Equal(t, key, readFile(t, out.Key))
}

// Verifies: CLI-0035
func TestWithFilesTheCommandRenewsWhenTheKeyFileDoesNotAgree(t *testing.T) {
	l, out := newLab(t), filesIn(t)
	_, err := l.runFiles(t, out, serverConfig().Config)
	require.NoError(t, err)

	other, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	otherPEM, err := ca.EncodeKey(other)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(out.Key, otherPEM, 0o600))

	issued, err := l.runFiles(t, out, serverConfig().Config)

	require.NoError(t, err)
	assert.True(t, issued)
}

// Verifies: CLI-0037
func TestWithFilesAFailedIssueChangesNoFile(t *testing.T) {
	l, out := newLab(t), filesIn(t)
	cfg := serverConfig().Config
	_, err := l.runFiles(t, out, cfg)
	require.NoError(t, err)

	cert, key := readFile(t, out.Cert), readFile(t, out.Key)
	cfg.CommonName = "wattle"

	l.register.Refuse(http.StatusBadRequest, "The O must be the brand name or the legal entity name.")

	_, err = l.runFiles(t, out, cfg)

	require.Error(t, err)
	assert.Equal(t, cert, readFile(t, out.Cert))
	assert.Equal(t, key, readFile(t, out.Key))
}

func userCSR(t *testing.T, kind ca.Kind, cn, o string, dns []string) ([]byte, *rsa.PrivateKey) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	csr, err := ca.NewCSR(key, kind, cn, o, dns)
	require.NoError(t, err)

	return csr, key
}

func csrConfig(csr []byte) certissue.Config {
	return certissue.Config{
		Kind: ca.Server, LegalEntityID: "le-1", RecordID: "brand-1", CSR: csr,
		RenewBefore: certcheck.DefaultRenewBefore, Now: serverConfig().Now,
	}
}

// Verifies: CLI-0127
func TestWithACSRTheCommandSendsThatCSRAndWritesOnlyTheCertificateFile(t *testing.T) {
	l, out := newLab(t), filesIn(t)
	out.Key = ""
	csr, key := userCSR(t, ca.Server, "wattle-bank", "Wattle Bank Ltd", []string{"wattle-bank", "localhost"})

	issued, err := l.runFiles(t, out, csrConfig(csr))
	require.NoError(t, err)
	assert.True(t, issued)

	assert.Equal(t, csr, l.register.LastCSR())

	chain, err := ca.ReadCertificates(out.Cert)
	require.NoError(t, err)
	assert.True(t, sameKey(chain[0].PublicKey, key.Public()))
	assert.Equal(t, "wattle-bank", chain[0].Subject.CommonName)

	entries, err := os.ReadDir(filepath.Dir(out.Cert))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

// Verifies: CLI-0129
func TestWithACSRTheKeyOfTheRenewalCheckIsTheKeyOfTheCSR(t *testing.T) {
	l, out := newLab(t), filesIn(t)
	out.Key = ""
	csr, _ := userCSR(t, ca.Server, "wattle-bank", "Wattle Bank Ltd", []string{"wattle-bank"})

	_, err := l.runFiles(t, out, csrConfig(csr))
	require.NoError(t, err)

	issued, err := l.runFiles(t, out, csrConfig(csr))
	require.NoError(t, err)
	assert.False(t, issued)

	otherCSR, _ := userCSR(t, ca.Server, "wattle-bank", "Wattle Bank Ltd", []string{"wattle-bank"})

	issued, err = l.runFiles(t, out, csrConfig(otherCSR))
	require.NoError(t, err)
	assert.True(t, issued)
}

func TestACSRThatIsNotOnePEMCSRIsAnErrorBeforeTheIssue(t *testing.T) {
	l, out := newLab(t), filesIn(t)
	out.Key = ""

	_, err := l.runFiles(t, out, csrConfig([]byte("not a CSR")))

	require.Error(t, err)
	assert.Empty(t, l.register.Issues())
}
