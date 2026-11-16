package main

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCRLURL = "http://ca.example/crl/intermediate.crl"

// verifyAt is one day after the test time, as Unix seconds for openssl verify -attime.
const verifyAt = "1790985600"

var testConfig = config{now: func() time.Time { return time.Date(2026, 10, 2, 6, 30, 0, 0, time.UTC) }, caKeyBits: 2048}

func pki(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()

	var out, errOut bytes.Buffer

	code = run(args, &out, &errOut, testConfig)

	return code, out.String(), errOut.String()
}

func openssl(t *testing.T, args ...string) (string, error) {
	t.Helper()

	_, err := exec.LookPath("openssl")
	require.NoError(t, err, "the tests need openssl")
	//nolint:gosec // the tests make all the arguments
	out, err := exec.CommandContext(t.Context(), "openssl", args...).CombinedOutput()

	return string(out), err
}

func newLabDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "pki")
	code, _, stderr := pki(t, "init", "-dir", dir)
	require.Zero(t, code, stderr)

	return dir
}

func opensslCSR(t *testing.T, subject string, extra ...string) string {
	t.Helper()
	dir := t.TempDir()
	csr := filepath.Join(dir, "request.csr")
	args := append(
		[]string{
			"req",
			"-new",
			"-newkey",
			"rsa:2048",
			"-nodes",
			"-sha256",
			"-keyout",
			filepath.Join(dir, "request.key"),
			"-subj",
			subject,
			"-out",
			csr,
		},
		extra...)
	out, err := openssl(t, args...)
	require.NoError(t, err, out)

	return csr
}

func TestInitCreatesTheLabCAsOneTime(t *testing.T) {
	dir := newLabDir(t)
	before, err := fs.ReadFile(os.DirFS(dir), "cdr/root.crt")
	require.NoError(t, err)

	code, stdout, _ := pki(t, "init", "-dir", dir)

	require.Zero(t, code)
	assert.Contains(t, stdout, "exist")

	after, err := fs.ReadFile(os.DirFS(dir), "cdr/root.crt")
	require.NoError(t, err)
	assert.Equal(t, before, after)
}

func TestInitFailsWhenOnlySomeLabCAFilesExist(t *testing.T) {
	dir := newLabDir(t)
	require.NoError(t, os.Remove(filepath.Join(dir, "cdr/intermediate.key")))

	code, _, stderr := pki(t, "init", "-dir", dir)

	assert.Equal(t, 1, code)
	assert.Contains(t, stderr, "cdr/intermediate.key is missing")
}

// Verifies: PKI-0002, PKI-0013, CA-0009
func TestIssueGivesCertificatesThatOpenSSLVerifiesAgainstTheCDRChainOnly(t *testing.T) {
	dir := newLabDir(t)

	tests := map[string]struct {
		kind, subject, purpose string
		extra                  []string
	}{
		"server": {
			"server", "/C=AU/O=Holder Bank/OU=Consumer Data Right/CN=mtls.holder.example", "sslserver",
			[]string{"-addext", "subjectAltName=DNS:api.holder.example"},
		},
		"client": {"client", "/C=AU/O=Recipient App/OU=Consumer Data Right/CN=Recipient App Web", "sslclient", nil},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cert := filepath.Join(t.TempDir(), "issued.crt")

			code, _, stderr := pki(
				t,
				"issue",
				"-dir",
				dir,
				"-crl-url",
				testCRLURL,
				"-type",
				tt.kind,
				"-out",
				cert,
				opensslCSR(t, tt.subject, tt.extra...),
			)

			require.Zero(t, code, stderr)
			out, err := openssl(
				t,
				"verify",
				"-attime",
				verifyAt,
				"-purpose",
				tt.purpose,
				"-CAfile",
				filepath.Join(dir, "cdr/ca-bundle.pem"),
				cert,
			)
			require.NoError(t, err, out)
			out, err = openssl(
				t,
				"verify",
				"-attime",
				verifyAt,
				"-purpose",
				tt.purpose,
				"-CAfile",
				filepath.Join(dir, "cdr/root.crt"),
				"-untrusted",
				filepath.Join(dir, "cdr/intermediate.crt"),
				cert,
			)
			require.NoError(t, err, out)
			out, err = openssl(t, "verify", "-attime", verifyAt, "-CAfile", filepath.Join(dir, "public/root.crt"), cert)
			assert.Error(t, err, out)
		})
	}
}

// Verifies: CA-0010
func TestIssueRefusesACSRThatBreaksTheCDSProfile(t *testing.T) {
	dir := newLabDir(t)
	cert := filepath.Join(t.TempDir(), "issued.crt")
	csr := opensslCSR(t, "/C=AU/O=Recipient App/CN=Recipient App Web")

	code, _, stderr := pki(t, "issue", "-dir", dir, "-crl-url", testCRLURL, "-type", "client", "-out", cert, csr)

	assert.Equal(t, 1, code)
	assert.Contains(t, stderr, `subject OU must be "Consumer Data Right"`)
	assert.NoFileExists(t, cert)
}

// Verifies: CA-0018, CA-0019
func TestIssueRefusesAServerCSRWithANameThatIsNotAHostName(t *testing.T) {
	dir := newLabDir(t)
	cert := filepath.Join(t.TempDir(), "issued.crt")
	csr := opensslCSR(
		t,
		"/C=AU/O=Holder Bank/OU=Consumer Data Right/CN=10.0.0.1",
		"-addext",
		"subjectAltName=DNS:*.holder.example",
	)

	code, _, stderr := pki(t, "issue", "-dir", dir, "-crl-url", testCRLURL, "-type", "server", "-out", cert, csr)

	assert.Equal(t, 1, code)
	assert.Contains(t, stderr, `subject CN "10.0.0.1" is an IP address, not a host name`)
	assert.Contains(t, stderr, `DNS name "*.holder.example" is not a host name`)
	assert.NoFileExists(t, cert)
}

// Verifies: PKI-0077, CA-0009
func TestIssuePublicGivesAKeyAndACertificateThatChainOnlyToThePublicRoot(t *testing.T) {
	dir := newLabDir(t)
	out := t.TempDir()
	cert, key := filepath.Join(out, "register.crt"), filepath.Join(out, "register.key")

	code, _, stderr := pki(t, "issue-public", "-dir", dir, "-cert", cert, "-key", key, "register", "localhost")

	require.Zero(t, code, stderr)
	result, err := openssl(
		t,
		"verify",
		"-attime",
		verifyAt,
		"-purpose",
		"sslserver",
		"-verify_hostname",
		"localhost",
		"-CAfile",
		filepath.Join(dir, "public/root.crt"),
		cert,
	)
	require.NoError(t, err, result)
	result, err = openssl(t, "verify", "-attime", verifyAt, "-CAfile", filepath.Join(dir, "cdr/ca-bundle.pem"), cert)
	require.Error(t, err, result)
	info, err := os.Stat(key)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm(), "lab containers read the key as another user")
}

func publicOnlyLabDir(t *testing.T, without ...string) string {
	t.Helper()
	dir := newLabDir(t)
	require.NoError(t, os.RemoveAll(filepath.Join(dir, "cdr")))

	for _, name := range without {
		require.NoError(t, os.Remove(filepath.Join(dir, name)))
	}

	return dir
}

func TestIssuePublicNeedsOnlyThePublicCA(t *testing.T) {
	dir := publicOnlyLabDir(t)
	out := t.TempDir()
	cert, key := filepath.Join(out, "wattle-bank.crt"), filepath.Join(out, "wattle-bank.key")
	names := []string{"wattle-bank.lab.example", "wattle-bank.cdr-lab.svc.cluster.local"}

	code, _, stderr := pki(t, append([]string{"issue-public", "-dir", dir, "-cert", cert, "-key", key}, names...)...)

	require.Zero(t, code, stderr)

	for _, name := range names {
		result, err := openssl(
			t, "verify", "-attime", verifyAt, "-purpose", "sslserver", "-verify_hostname", name,
			"-CAfile", filepath.Join(dir, "public/root.crt"), cert,
		)
		require.NoError(t, err, result)
	}
}

func TestIssuePublicNamesTheMissingPublicCAFile(t *testing.T) {
	for _, missing := range []string{"public/root.crt", "public/root.key"} {
		t.Run(missing, func(t *testing.T) {
			dir := publicOnlyLabDir(t, missing)
			out := t.TempDir()

			code, _, stderr := pki(
				t, "issue-public", "-dir", dir,
				"-cert", filepath.Join(out, "a.crt"), "-key", filepath.Join(out, "a.key"), "localhost",
			)

			assert.Equal(t, 1, code)
			assert.Contains(t, stderr, filepath.Join(dir, missing))
			assert.NotContains(t, stderr, filepath.Join(dir, "cdr"))
		})
	}
}

func TestAnUnknownCommandGivesTheUsage(t *testing.T) {
	code, _, stderr := pki(t, "sign")

	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "usage: pki")
}

// The Register keeps the only list of revocations, and signs each CRL.
// Verifies: PKI-0067
func TestThePKICommandDoesNotRevokeOrSignACRL(t *testing.T) {
	dir := newLabDir(t)

	for _, command := range []string{"revoke", "crl"} {
		code, _, stderr := pki(t, command, "-dir", dir)

		assert.Equal(t, 2, code, command)
		assert.Contains(t, stderr, "usage: pki", command)
		assert.NotContains(t, stderr, command+" ", "the usage does not give the command")
	}

	assert.NoDirExists(t, filepath.Join(dir, "crl"))
}

func issueClient(t *testing.T, dir, name string) string {
	t.Helper()
	cert := filepath.Join(t.TempDir(), name+".crt")
	code, _, stderr := pki(t, "issue", "-dir", dir, "-crl-url", testCRLURL, "-type", "client", "-out", cert,
		opensslCSR(t, "/C=AU/O=Recipient App/OU=Consumer Data Right/CN="+name))
	require.Zero(t, code, stderr)

	return cert
}

// Verifies: CA-0034
func TestACDRCertificateGivesTheCRLURL(t *testing.T) {
	dir := newLabDir(t)
	cert := issueClient(t, dir, "Recipient App Web")

	out, err := openssl(t, "x509", "-in", cert, "-noout", "-ext", "crlDistributionPoints")

	require.NoError(t, err, out)
	assert.Contains(t, out, "URI:"+testCRLURL)
}

func TestEachCommandNeedsTheLabDirectory(t *testing.T) {
	for _, command := range []string{"init", "issue", "issue-public"} {
		t.Run(command, func(t *testing.T) {
			code, _, stderr := pki(t, command)

			assert.Equal(t, 1, code)
			assert.Contains(t, stderr, "-dir is required")
		})
	}
}

func TestIssueNeedsAnAbsoluteHTTPCRLURL(t *testing.T) {
	dir := newLabDir(t)
	csr := opensslCSR(t, "/C=AU/O=Recipient App/OU=Consumer Data Right/CN=Recipient App Web")

	tests := map[string]struct {
		args    []string
		wantErr string
	}{
		"no CRL URL":       {nil, "-crl-url is required"},
		"relative URL":     {[]string{"-crl-url", "/crl/intermediate.crl"}, "is not an absolute http or https URL"},
		"other URL scheme": {[]string{"-crl-url", "ldap://ca.example/crl"}, "is not an absolute http or https URL"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cert := filepath.Join(t.TempDir(), "issued.crt")
			args := append([]string{"issue", "-dir", dir, "-type", "client", "-out", cert}, tt.args...)

			code, _, stderr := pki(t, append(args, csr)...)

			assert.Equal(t, 1, code)
			assert.Contains(t, stderr, tt.wantErr)
			assert.NoFileExists(t, cert)
		})
	}
}
