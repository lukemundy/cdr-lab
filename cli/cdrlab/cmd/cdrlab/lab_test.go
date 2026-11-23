package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/certlist"
	"github.com/lukemundy/cdr-lab/cli/cdrlab/internal/register/registertest"
)

type composeCall struct {
	dir    string
	env    []string
	argv   []string
	issues int
}

// fakeCompose records each call, and the number of Register issues at the time of the call.
type fakeCompose struct {
	register *registertest.Register
	calls    []composeCall
	failOn   string
	ps       string
}

func (f *fakeCompose) exec(_ context.Context, dir string, env []string, stdout, _ io.Writer, argv []string) error {
	f.calls = append(f.calls, composeCall{dir: dir, env: env, argv: argv, issues: len(f.register.Issues())})

	if slices.Contains(argv, "ps") {
		_, _ = io.WriteString(stdout, f.ps)
		return nil
	}

	if f.failOn != "" && argv[len(argv)-1] == f.failOn {
		return errors.New("exit status 1")
	}

	return nil
}

func (f *fakeCompose) upCalls() []composeCall {
	var ups []composeCall

	for _, c := range f.calls {
		if slices.Contains(c.argv, "up") {
			ups = append(ups, c)
		}
	}

	return ups
}

func envValue(env []string, key string) (string, bool) {
	for _, kv := range slices.Backward(env) {
		if k, v, _ := strings.Cut(kv, "="); k == key {
			return v, true
		}
	}

	return "", false
}

// newLab writes a compose project with a certificate list for the fake Register.
func newLab(t *testing.T, h *harness) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte("services: {}\n"), 0o600))

	list := certlist.List{
		Register: certlist.Register{URL: h.register.URL, OperatorToken: registertest.Token},
		Public: []certlist.Public{
			{Cert: ".lab/pki/register/public.crt", Key: ".lab/pki/register/public.key", Names: []string{"register"}},
		},
		CDR: []certlist.CDR{
			{
				Kind: "server", LegalEntityID: "le-1", BrandID: "brand-1", CN: "wattle-bank", O: "Wattle Bank Ltd",
				DNS:  []string{"wattle-bank", "localhost"},
				Cert: ".lab/pki/wattle-bank/secure.crt", Key: ".lab/pki/wattle-bank/secure.key",
			},
			{
				Kind:              "client",
				LegalEntityID:     "le-2",
				SoftwareProductID: "sp-2",
				CN:                "Token Test Active",
				O:                 "Token Test",
				Cert:              ".lab/pki/conformance/client.crt",
				Key:               ".lab/pki/conformance/client.key",
			},
		},
	}

	data, err := json.Marshal(list)
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lab"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, certlist.File), data, 0o600))

	return dir
}

func readFiles(t *testing.T, dir string) map[string]string {
	t.Helper()

	files := map[string]string{}
	err := filepath.WalkDir(filepath.Join(dir, ".lab/pki"), func(path string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}

		data, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
		files[path] = string(data)

		return err
	})
	require.NoError(t, err)

	return files
}

// Verifies: CLI-0108, CLI-0135, CLI-0137, CLI-0146, CLI-0148, CLI-0149
func TestUpMakesTheCAsAndTheCertificatesAndStartsTheRegisterBeforeTheCDRCertificates(t *testing.T) {
	h := newHarness(t)
	dir := newLab(t, h)

	code, stdout, stderr := h.run("up", "-source", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	assert.FileExists(t, filepath.Join(dir, ".lab/pki/cdr/ca-bundle.pem"))
	assert.FileExists(t, filepath.Join(dir, ".lab/pki/public/root.crt"))
	assert.FileExists(t, filepath.Join(dir, ".lab/pki/register/public.crt"))
	assert.Equal(t, os.FileMode(0o644), fileMode(t, filepath.Join(dir, ".lab/pki/register/public.key")))
	assert.FileExists(t, filepath.Join(dir, ".lab/pki/wattle-bank/secure.crt"))
	assert.Equal(t, os.FileMode(0o644), fileMode(t, filepath.Join(dir, ".lab/pki/wattle-bank/secure.key")))
	assert.Equal(t, os.FileMode(0o644), fileMode(t, filepath.Join(dir, ".lab/pki/conformance/client.key")))
	assert.Len(t, h.register.Issues(), 2)
	assert.NotContains(t, stdout, registertest.Token)

	ups := h.compose.upCalls()
	require.Len(t, ups, 2)
	assert.Equal(t, dir, ups[0].dir)
	assert.Equal(t, "register", ups[0].argv[len(ups[0].argv)-1])
	assert.Zero(t, ups[0].issues)
	assert.Equal(t, 2, ups[1].issues)
	assert.Equal(t, []string{"docker", "compose", "up", "--detach", "--build", "--wait", "--wait-timeout", "60"},
		ups[1].argv)
}

// Verifies: CLI-0062, CLI-0121
func TestUpAgainWithNoChangeChangesNoCertificate(t *testing.T) {
	h := newHarness(t)
	dir := newLab(t, h)

	code, _, stderr := h.run("up", "-source", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	before := readFiles(t, dir)

	code, stdout, stderr := h.run("up", "-source", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	assert.Equal(t, before, readFiles(t, dir))
	assert.Len(t, h.register.Issues(), 2)
	assert.NotContains(t, stdout, "wrote")
}

// Verifies: CLI-0150, CLI-0138, CLI-0153
func TestUpStartsOnlyTheNamedServicesWithTheImagesThatAreThereAndTheWaitTime(t *testing.T) {
	h := newHarness(t)
	dir := newLab(t, h)

	code, _, stderr := h.run("up", "-source", "-dir", dir, "-no-build", "-wait-timeout", "2m", "wattle-bank", "mail")
	require.Equal(t, 0, code, stderr)

	ups := h.compose.upCalls()
	require.Len(t, ups, 2)
	assert.Equal(t, []string{
		"docker", "compose", "up", "--detach", "--no-build", "--wait", "--wait-timeout", "120",
		"register",
	}, ups[0].argv)
	assert.Equal(t, []string{
		"docker", "compose", "up", "--detach", "--no-build", "--wait", "--wait-timeout", "120",
		"wattle-bank", "mail",
	}, ups[1].argv)
}

// Verifies: CLI-0151, CLI-0152
func TestUpSetsTheProfilesAndTheTelemetryOfCompose(t *testing.T) {
	cases := []struct {
		flags    []string
		profiles string
		otlpOff  bool
	}{
		{nil, "telemetry", false},
		{[]string{"-all-banks"}, "telemetry,all-banks", false},
		{[]string{"-no-telemetry"}, "", true},
		{[]string{"-no-telemetry", "-all-banks"}, "all-banks", true},
	}

	for _, c := range cases {
		t.Run(strings.Join(c.flags, " "), func(t *testing.T) {
			h := newHarness(t)
			h.environ = []string{"COMPOSE_PROFILES=other", "LAB_OTLP_ENDPOINT=http://elsewhere:4318", "HOME=/home/lab"}
			dir := newLab(t, h)

			code, _, stderr := h.run(append([]string{"up", "-source", "-dir", dir}, c.flags...)...)
			require.Equal(t, 0, code, stderr)

			for _, call := range h.compose.upCalls() {
				profiles, _ := envValue(call.env, "COMPOSE_PROFILES")
				assert.Equal(t, c.profiles, profiles)

				otlp, _ := envValue(call.env, "LAB_OTLP_ENDPOINT")
				assert.Equal(t, c.otlpOff, otlp == "")

				home, _ := envValue(call.env, "HOME")
				assert.Equal(t, "/home/lab", home)
			}
		})
	}
}

// Verifies: CLI-0136
func TestUpRunsTheComposeCommandOfTheVariable(t *testing.T) {
	h := newHarness(t)
	h.env["CDRLAB_COMPOSE"] = "docker-compose"
	dir := newLab(t, h)

	code, _, stderr := h.run("up", "-source", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	assert.Equal(t, "docker-compose", h.compose.upCalls()[0].argv[0])
}

// Verifies: CLI-0146, CLI-0153
func TestUpStopsWhenTheRegisterIsNotHealthyAndNamesTheService(t *testing.T) {
	h := newHarness(t)
	h.compose.failOn = "register"
	h.compose.ps = `{"Service":"register","State":"running","Health":"unhealthy","ExitCode":0}` + "\n"
	dir := newLab(t, h)

	code, _, stderr := h.run("up", "-source", "-dir", dir)

	assert.Equal(t, 1, code)
	assert.Contains(t, stderr, "not healthy: register")
	assert.Empty(t, h.register.Issues())
	assert.Len(t, h.compose.upCalls(), 1)
}

func TestUpWithNoSourceFlagTellsTheUserToUseTheSourceMode(t *testing.T) {
	h := newHarness(t)

	code, _, stderr := h.run("up", "-dir", newLab(t, h))

	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "-source")
	assert.Empty(t, h.compose.calls)
}

func TestUpFailsWithNoCertificateList(t *testing.T) {
	h := newHarness(t)
	dir := t.TempDir()

	code, _, stderr := h.run("up", "-source", "-dir", dir)

	assert.Equal(t, 1, code)
	assert.Contains(t, stderr, certlist.File)
	assert.Empty(t, h.compose.calls)
}

// Verifies: CLI-0109, CLI-0154
func TestDownStopsTheLabAndKeepsTheCertificates(t *testing.T) {
	h := newHarness(t)
	dir := newLab(t, h)
	code, _, stderr := h.run("up", "-source", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	code, _, stderr = h.run("down", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	last := h.compose.calls[len(h.compose.calls)-1]
	assert.Equal(t, []string{"docker", "compose", "--profile", "*", "down", "--remove-orphans"}, last.argv)
	assert.Equal(t, dir, last.dir)
	assert.DirExists(t, filepath.Join(dir, ".lab/pki"))
}

// Verifies: CLI-0110, CLI-0155
func TestResetStopsTheLabAndDeletesItsVolumesAndCertificates(t *testing.T) {
	h := newHarness(t)
	dir := newLab(t, h)
	code, _, stderr := h.run("up", "-source", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	code, _, stderr = h.run("reset", "-dir", dir)
	require.Equal(t, 0, code, stderr)

	last := h.compose.calls[len(h.compose.calls)-1]
	assert.Equal(t, []string{"docker", "compose", "--profile", "*", "down", "--remove-orphans", "--volumes"},
		last.argv)
	assert.NoDirExists(t, filepath.Join(dir, ".lab/pki"))
	assert.FileExists(t, filepath.Join(dir, "compose.yaml"))
}

// Verifies: CLI-0134, CLI-0157
func TestDownAndResetInADirectoryWithNoLabChangeNothing(t *testing.T) {
	for _, command := range []string{"down", "reset"} {
		t.Run(command, func(t *testing.T) {
			h := newHarness(t)
			dir := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Join(dir, ".lab/pki"), 0o750))

			code, stdout, stderr := h.run(command, "-dir", dir)
			require.Equal(t, 0, code, stderr)

			assert.Contains(t, stdout, "no change")
			assert.Empty(t, h.compose.calls)
			assert.DirExists(t, filepath.Join(dir, ".lab/pki"))
		})
	}
}

func TestResetStopsAtAComposeFailureAndKeepsTheCertificates(t *testing.T) {
	h := newHarness(t)
	h.compose.failOn = "--volumes"
	dir := newLab(t, h)
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".lab/pki"), 0o750))

	code, _, _ := h.run("reset", "-dir", dir)

	assert.Equal(t, 1, code)
	assert.DirExists(t, filepath.Join(dir, ".lab/pki"))
}
