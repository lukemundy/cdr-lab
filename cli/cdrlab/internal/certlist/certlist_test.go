package certlist_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certlist"
)

func writeList(t *testing.T, body string) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lab"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, certlist.File), []byte(body), 0o600))

	return dir
}

const valid = `{
  "register": {"url": "http://127.0.0.1:8080", "operator-token": "token"},
  "public": [{"cert": ".lab/pki/register/public.crt", "key": ".lab/pki/register/public.key", "names": ["register"]}],
  "cdr": [
    {"kind": "server", "legal-entity-id": "le-1", "brand-id": "b-1", "cn": "wattle-bank", "o": "Wattle Bank",
     "dns": ["wattle-bank", "localhost"],
     "cert": ".lab/pki/wattle-bank/secure.crt", "key": ".lab/pki/wattle-bank/secure.key"},
    {"kind": "client", "legal-entity-id": "le-2", "software-product-id": "sp-2", "cn": "Client", "o": "Recipient",
     "cert": ".lab/pki/conformance/client.crt", "key": ".lab/pki/conformance/client.key"}
  ]
}`

// Verifies: CLI-0147, CLI-0193
func TestLoadReadsTheListOfTheLabDirectory(t *testing.T) {
	list, err := certlist.Load(writeList(t, valid))
	require.NoError(t, err)

	assert.Equal(t, "http://127.0.0.1:8080", list.Register.URL)
	assert.Equal(t, "token", list.Register.OperatorToken)
	require.Len(t, list.Public, 1)
	assert.Equal(t, []string{"register"}, list.Public[0].Names)
	require.Len(t, list.CDR, 2)
	assert.Equal(t, []string{"wattle-bank", "localhost"}, list.CDR[0].DNS)
	assert.Equal(t, "sp-2", list.CDR[1].SoftwareProductID)
}

// Verifies: CLI-0194
func TestFlagsGivesTheKeysOfACDRCertificateAsFlagsInTheLabDirectory(t *testing.T) {
	list, err := certlist.Load(writeList(t, valid))
	require.NoError(t, err)

	assert.Equal(t, []string{
		"-kind", "server", "-legal-entity-id", "le-1", "-brand-id", "b-1", "-cn", "wattle-bank", "-o", "Wattle Bank",
		"-dns", "wattle-bank", "-dns", "localhost",
		"-cert", filepath.Join("lab-dir", ".lab/pki/wattle-bank/secure.crt"),
		"-key", filepath.Join("lab-dir", ".lab/pki/wattle-bank/secure.key"),
	}, list.CDR[0].Flags("lab-dir"))
	assert.Equal(t, []string{
		"-kind",
		"client",
		"-legal-entity-id",
		"le-2",
		"-software-product-id",
		"sp-2",
		"-cn",
		"Client",
		"-o",
		"Recipient",
		"-cert",
		filepath.Join("lab-dir", ".lab/pki/conformance/client.crt"),
		"-key",
		filepath.Join("lab-dir", ".lab/pki/conformance/client.key"),
	}, list.CDR[1].Flags("lab-dir"))
}

func TestLoadFailsForAListThatIsAbsent(t *testing.T) {
	_, err := certlist.Load(t.TempDir())

	require.Error(t, err)
	assert.Contains(t, err.Error(), certlist.File)
}

func TestLoadRefusesAKeyThatIsNotKnown(t *testing.T) {
	_, err := certlist.Load(writeList(t, `{"register": {"url": "u", "operator-token": "t"}, "extra": 1}`))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "extra")
}

func TestLoadRefusesAListWithNoRegisterURLOrToken(t *testing.T) {
	_, err := certlist.Load(writeList(t, `{"register": {"url": "u"}}`))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operator-token")
}

// Verifies: CLI-0193
func TestLoadRefusesAPathOutsideTheLabDirectory(t *testing.T) {
	for _, path := range []string{"../outside.crt", "/etc/outside.crt", ""} {
		t.Run(path, func(t *testing.T) {
			_, err := certlist.Load(writeList(t, `{"register": {"url": "u", "operator-token": "t"},
				"public": [{"cert": "`+path+`", "key": "k", "names": ["n"]}]}`))

			require.Error(t, err)
			assert.Contains(t, err.Error(), "lab directory")
		})
	}
}

func TestLoadRefusesAPublicCertificateWithNoName(t *testing.T) {
	_, err := certlist.Load(writeList(t, `{"register": {"url": "u", "operator-token": "t"},
		"public": [{"cert": "c", "key": "k", "names": []}]}`))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "names")
}

// The committed list of the compose project must load, so that a bad edit fails the tests.
func TestTheListOfTheRepositoryLoads(t *testing.T) {
	list, err := certlist.Load("../../../..")
	require.NoError(t, err)

	assert.NotEmpty(t, list.Public)
	assert.NotEmpty(t, list.CDR)
}
