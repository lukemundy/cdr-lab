package main

import (
	"crypto/rand"
	"crypto/rsa"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

func fileMode(t *testing.T, path string) os.FileMode {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)

	return info.Mode().Perm()
}

// Verifies: CLI-0105, CLI-0113, CLI-0119
func TestCAInitWithADirectoryWritesTheFilesAndThenChangesNothing(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()

	code, stdout, stderr := h.run("ca", "init", "-dir", dir)
	require.Equal(t, 0, code, stderr)
	assert.Contains(t, stdout, "created")
	assert.FileExists(t, filepath.Join(dir, "cdr/ca-bundle.pem"))
	assert.NoFileExists(t, filepath.Join(dir, "cdr/root.key"))

	code, stdout, _ = h.run("ca", "init", "-dir", dir)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "no change")
	assert.Zero(t, h.kubeCalls)
}

func TestCAInitRefusesADirectoryAndASecret(t *testing.T) {
	code, _, stderr := newHarness(t).run("ca", "init", "-dir", t.TempDir(), "-intermediate-secret", "i")

	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "-intermediate-secret")
}

func (h *harness) fileIssueArgs(dir string, extra ...string) []string {
	return append([]string{
		"cert", "issue", "-register-url", h.register.URL, "-kind", "server",
		"-legal-entity-id", "le-1", "-brand-id", "brand-1", "-cn", "wattle-bank", "-o", "Wattle Bank Ltd",
		"-dns", "localhost",
		"-cert", filepath.Join(dir, "wattle-bank/secure.crt"), "-key", filepath.Join(dir, "wattle-bank/secure.key"),
	}, extra...)
}

// Verifies: CLI-0106, CLI-0113, CLI-0124, CLI-0125
func TestCertIssueWithFilesWritesTheCertificateAndAKeyForTheOwnerAndMakesNoKubernetesCall(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()

	code, stdout, stderr := h.run(h.fileIssueArgs(dir)...)
	require.Equal(t, 0, code, stderr)
	assert.Contains(t, stdout, "valid until")
	assert.Equal(t, os.FileMode(0o600), fileMode(t, filepath.Join(dir, "wattle-bank/secure.key")))

	code, stdout, _ = h.run(h.fileIssueArgs(dir)...)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "no change")
	assert.Len(t, h.register.Issues(), 1)
	assert.Zero(t, h.kubeCalls)
}

// Verifies: CLI-0125
func TestCertIssueWithFilesLetsAnyUserReadTheKeyWhenAnInputPermitsIt(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()

	code, _, stderr := h.run(h.fileIssueArgs(dir, "-key-readable-by-all")...)
	require.Equal(t, 0, code, stderr)

	assert.Equal(t, os.FileMode(0o644), fileMode(t, filepath.Join(dir, "wattle-bank/secure.key")))
}

func writeCSR(t *testing.T, dir string) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	csr, err := ca.NewCSR(key, ca.Client, "Token Test Active", "Token Test", nil)
	require.NoError(t, err)

	path := filepath.Join(dir, "client.csr")
	require.NoError(t, os.WriteFile(path, csr, 0o600))

	return path
}

// Verifies: CLI-0127
func TestCertIssueWithACSRFileSendsItAndWritesOnlyTheCertificate(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()
	csrPath := writeCSR(t, dir)

	code, _, stderr := h.run("cert", "issue", "-register-url", h.register.URL, "-kind", "client",
		"-legal-entity-id", "le-2", "-software-product-id", "sp-2",
		"-csr", csrPath, "-cert", filepath.Join(dir, "client.crt"))
	require.Equal(t, 0, code, stderr)

	csr, err := fs.ReadFile(os.DirFS(dir), "client.csr")
	require.NoError(t, err)
	assert.Equal(t, csr, h.register.LastCSR())
	assert.FileExists(t, filepath.Join(dir, "client.crt"))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

// Verifies: CLI-0128
func TestCertIssueRefusesACSRFileWithASubjectADNSNameOrAKeyFile(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()
	csrPath := writeCSR(t, dir)
	base := []string{
		"cert", "issue", "-register-url", h.register.URL, "-kind", "server",
		"-legal-entity-id", "le-1", "-brand-id", "brand-1", "-csr", csrPath, "-cert", filepath.Join(dir, "a.crt"),
	}

	for name, extra := range map[string][]string{
		"a CN":       {"-cn", "wattle-bank"},
		"an O":       {"-o", "Wattle Bank Ltd"},
		"a DNS name": {"-dns", "localhost"},
		"a key file": {"-key", filepath.Join(dir, "a.key")},
	} {
		t.Run(name, func(t *testing.T) {
			code, _, stderr := h.run(append(append([]string{}, base...), extra...)...)

			assert.Equal(t, 2, code)
			assert.Contains(t, stderr, "-csr cannot go with")
		})
	}

	assert.Empty(t, h.register.Issues())
}

func TestCertIssueRefusesTwoOutputsOrNone(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()

	server := []string{
		"cert", "issue", "-register-url", h.register.URL, "-kind", "server",
		"-legal-entity-id", "le-1", "-brand-id", "brand-1", "-cn", "wattle-bank", "-o", "Wattle Bank Ltd",
	}
	with := func(args ...string) []string { return append(append([]string{}, server...), args...) }
	certFile, keyFile := filepath.Join(dir, "a.crt"), filepath.Join(dir, "a.key")

	for name, args := range map[string][]string{
		"a Secret and files":                  with("-secret", "s", "-cert", certFile, "-key", keyFile),
		"no output":                           with(),
		"a certificate file with no key file": with("-cert", certFile),
		"a Secret and a CSR":                  with("-secret", "s", "-csr", filepath.Join(dir, "a.csr")),
	} {
		t.Run(name, func(t *testing.T) {
			code, _, stderr := h.run(args...)

			assert.Equal(t, 2, code)
			assert.NotEmpty(t, stderr)
		})
	}

	assert.Empty(t, h.register.Issues())
}

// Verifies: CLI-0107, CLI-0130, CLI-0133
func TestPublicCertIssueWritesTheCertificateAndThenChangesNothing(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()
	code, _, stderr := h.run("ca", "init", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	args := []string{
		"public-cert", "issue", "-dir", dir,
		"-cert", filepath.Join(dir, "register/public.crt"), "-key", filepath.Join(dir, "register/public.key"),
		"register", "localhost",
	}

	code, stdout, stderr := h.run(args...)
	require.Equal(t, 0, code, stderr)
	assert.Contains(t, stdout, "valid until")
	assert.Equal(t, os.FileMode(0o600), fileMode(t, filepath.Join(dir, "register/public.key")))

	certs, err := ca.ReadCertificates(filepath.Join(dir, "register/public.crt"))
	require.NoError(t, err)
	assert.Equal(t, "register", certs[0].Subject.CommonName)

	code, stdout, _ = h.run(args...)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "no change")
	assert.Zero(t, h.kubeCalls)
	assert.Empty(t, h.register.Issues())
}

func TestPublicCertIssueNeedsADNSName(t *testing.T) {
	dir := t.TempDir()

	code, _, stderr := newHarness(t).run("public-cert", "issue", "-dir", dir,
		"-cert", filepath.Join(dir, "a.crt"), "-key", filepath.Join(dir, "a.key"))

	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "DNS name")
}
