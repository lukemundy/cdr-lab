package main

import (
	"bytes"
	"context"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/kube/kubetest"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register/registertest"
)

type harness struct {
	api       *kubetest.API
	register  *registertest.Register
	env       map[string]string
	environ   []string
	compose   *fakeCompose
	kubeCalls int
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	register := registertest.New(t)

	return &harness{
		api:      kubetest.New(t),
		register: register,
		env:      map[string]string{"REGISTER_OPERATOR_TOKEN": registertest.Token},
		compose:  &fakeCompose{register: register},
	}
}

func (h *harness) run(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer

	code = run(context.Background(), args, &out, &errOut, deps{
		secrets: func() (*kube.Client, error) {
			h.kubeCalls++

			return kube.New(h.api.URL, kubetest.Namespace, h.api.TokenFile, h.api.Client()), nil
		},
		httpClient: http.DefaultClient,
		getenv:     func(k string) string { return h.env[k] },
		environ:    func() []string { return h.environ },
		compose:    h.compose.exec,
		now:        time.Now,
		caKeyBits:  2048,
	})

	return code, out.String(), errOut.String()
}

func (h *harness) issueArgs(extra ...string) []string {
	return append([]string{
		"cert",
		"issue",
		"-register-url",
		h.register.URL,
		"-kind",
		"server",
		"-legal-entity-id",
		"le-1",
		"-brand-id",
		"brand-1",
		"-cn",
		"wattle-bank",
		"-o",
		"Wattle Bank Ltd",
		"-dns",
		"wattle-bank",
		"-dns",
		"localhost",
		"-secret",
		"wattle-bank-secure-tls",
	}, extra...)
}

func TestHelpGivesTheUsageAndExitsWithZero(t *testing.T) {
	code, stdout, _ := newHarness(t).run("-h")

	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "cdrlab ca init")
	assert.Contains(t, stdout, "cdrlab cert issue")
	assert.Contains(t, stdout, "cdrlab public-cert issue")
	assert.Contains(t, stdout, "cdrlab version")
}

// Verifies: CLI-0104, CLI-0114, CLI-0117, CLI-0118
func TestVersionGivesTheVersionTheCommitAndTheBuildTime(t *testing.T) {
	code, stdout, _ := newHarness(t).run("version")

	assert.Equal(t, 0, code)
	assert.Equal(t, "cdrlab dev, commit unknown, built unknown\n", stdout)
}

// Verifies: CLI-0003
func TestAnUnknownCommandFailsWithTheUsage(t *testing.T) {
	code, _, stderr := newHarness(t).run("pki")

	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "usage")
}

// Verifies: CLI-0010, CLI-0014
func TestCAInitCreatesTheSecretsAndThenChangesNothing(t *testing.T) {
	h := newHarness(t)
	args := []string{"ca", "init", "-intermediate-secret", "cdr-intermediate", "-ca-bundle-secret", "cdr-ca-bundle"}

	code, stdout, _ := h.run(args...)
	require.Equal(t, 0, code)
	assert.Contains(t, stdout, "created")
	assert.ElementsMatch(t, []string{"cdr-intermediate", "cdr-ca-bundle"}, h.api.Names())

	code, stdout, _ = h.run(args...)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "no change")
}

// Verifies: CLI-0003, CLI-0015
func TestCAInitFailsWithTheCauseOnTheErrorOutput(t *testing.T) {
	h := newHarness(t)
	h.api.Put("cdr-intermediate", kubetest.Secret{Type: kube.TLSType})

	code, _, stderr := h.run(
		"ca",
		"init",
		"-intermediate-secret",
		"cdr-intermediate",
		"-ca-bundle-secret",
		"cdr-ca-bundle",
	)

	assert.Equal(t, 1, code)
	assert.Contains(t, stderr, `"cdr-ca-bundle" is absent`)
}

func TestCAInitNeedsTheNamesOfTheSecrets(t *testing.T) {
	code, _, stderr := newHarness(t).run("ca", "init", "-intermediate-secret", "cdr-intermediate")

	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "-ca-bundle-secret")
}

// Verifies: CLI-0030, CLI-0031, CLI-0035
func TestCertIssueWritesTheSecretAndThenChangesNothing(t *testing.T) {
	h := newHarness(t)

	code, stdout, stderr := h.run(h.issueArgs()...)
	require.Equal(t, 0, code, stderr)
	assert.Contains(t, stdout, "wattle-bank-secure-tls")
	assert.Contains(t, stdout, "valid until")

	code, stdout, _ = h.run(h.issueArgs()...)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "no change")
	assert.Len(t, h.register.Issues(), 1)
}

// Verifies: CLI-0036
func TestCertIssueTakesARenewalWindow(t *testing.T) {
	h := newHarness(t)
	code, _, _ := h.run(h.issueArgs()...)
	require.Equal(t, 0, code)

	code, _, _ = h.run(h.issueArgs("-renew-before", "9600h")...)

	assert.Equal(t, 0, code)
	assert.Len(t, h.register.Issues(), 2)
}

func TestCertIssueTakesASoftwareProductForAClientCertificate(t *testing.T) {
	h := newHarness(t)

	code, _, stderr := h.run("cert", "issue", "-register-url", h.register.URL, "-kind", "client",
		"-legal-entity-id", "le-2", "-software-product-id", "sp-2", "-cn", "Token Test Active", "-o", "Token Test",
		"-secret", "conformance-client")

	require.Equal(t, 0, code, stderr)
	assert.Equal(
		t,
		[]string{"POST /v1/data-recipients/le-2/software-products/sp-2/certificates/issue"},
		h.register.Issues(),
	)
}

func TestCertIssueRefusesInputsThatDoNotAgreeWithTheKind(t *testing.T) {
	h := newHarness(t)

	server := []string{"cert", "issue", "-register-url", "http://r", "-legal-entity-id", "le", "-cn", "c", "-o", "o"}
	withSecret := func(args ...string) []string { return append(append(slices.Clone(server), args...), "-secret", "s") }

	for name, args := range map[string][]string{
		"no kind":                  withSecret("-brand-id", "b"),
		"a server with no brand":   withSecret("-kind", "server", "-software-product-id", "sp"),
		"a client with a DNS name": withSecret("-kind", "client", "-software-product-id", "sp", "-dns", "d"),
		"a client with a brand":    withSecret("-kind", "client", "-brand-id", "b", "-software-product-id", "sp"),
		"no Register URL": {
			"cert", "issue", "-kind", "client", "-legal-entity-id", "le", "-software-product-id", "sp",
			"-cn", "c", "-o", "o", "-secret", "s",
		},
		"a renewal window of 0": withSecret("-kind", "client", "-software-product-id", "sp", "-renew-before", "0s"),
	} {
		t.Run(name, func(t *testing.T) {
			code, _, stderr := h.run(args...)

			assert.Equal(t, 2, code)
			assert.NotEmpty(t, stderr)
		})
	}

	assert.Empty(t, h.register.Issues())
}

func TestCertIssueNeedsTheOperatorToken(t *testing.T) {
	h := newHarness(t)
	delete(h.env, "REGISTER_OPERATOR_TOKEN")

	code, _, stderr := h.run(h.issueArgs()...)

	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "REGISTER_OPERATOR_TOKEN")
}

// Verifies: CLI-0060
func TestTheOutputHasNoKeyAndNoToken(t *testing.T) {
	h := newHarness(t)
	h.env["REGISTER_OPERATOR_TOKEN"] = "not-the-token"

	_, failOut, failErr := h.run(h.issueArgs()...)

	h.env["REGISTER_OPERATOR_TOKEN"] = registertest.Token
	_, out, errOut := h.run(h.issueArgs()...)
	_, caOut, caErr := h.run("ca", "init", "-intermediate-secret", "i", "-ca-bundle-secret", "b")

	for _, text := range []string{failOut, failErr, out, errOut, caOut, caErr} {
		assert.NotContains(t, text, "PRIVATE KEY")
		assert.NotContains(t, text, "not-the-token")
		assert.NotContains(t, text, registertest.Token)
		assert.NotContains(t, text, kubetest.Token)
	}
}

// Verifies: CLI-0112
func TestNoCommandRevokesACertificateOrSignsACRL(t *testing.T) {
	h := newHarness(t)
	_, usageText, _ := h.run("-h")

	assert.NotContains(t, usageText, "revoke")
	assert.NotContains(t, usageText, "crl")

	for _, args := range [][]string{{"cert", "revoke"}, {"crl", "sign"}} {
		code, _, _ := h.run(args...)
		assert.Equal(t, 2, code, args)
	}
}
