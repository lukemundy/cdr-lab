package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testInputs = []string{"services/register", "go.mod", ":(exclude)*.md"}

type history struct {
	t    *testing.T
	repo gitRepo
}

func newHistory(t *testing.T) history {
	t.Helper()

	h := history{t: t, repo: gitRepo{dir: t.TempDir()}}
	h.git("init", "--quiet", "--initial-branch=main")

	return h
}

func (h history) git(args ...string) string {
	h.t.Helper()

	//nolint:gosec // The arguments are the git commands of the tests.
	cmd := exec.CommandContext(h.t.Context(), "git",
		append([]string{"-C", h.repo.dir, "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false"}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")

	out, err := cmd.CombinedOutput()
	require.NoError(h.t, err, string(out))

	return strings.TrimSpace(string(out))
}

// commit writes the file and gives the full commit ID.
func (h history) commit(path string) string {
	h.t.Helper()

	file := filepath.Join(h.repo.dir, path)
	require.NoError(h.t, os.MkdirAll(filepath.Dir(file), 0o750))
	require.NoError(h.t, os.WriteFile(file, []byte(h.t.Name()+path+h.git("rev-list", "--all", "--count")), 0o600))
	h.git("add", "--all")
	h.git("commit", "--quiet", "--message", "change "+path)

	return h.git("rev-parse", "HEAD")
}

// change writes the files, commits them with the message and gives the full commit ID. An empty content deletes the
// file.
func (h history) change(message string, files map[string]string) string {
	h.t.Helper()

	for path, content := range files {
		file := filepath.Join(h.repo.dir, path)
		if content == "" {
			require.NoError(h.t, os.Remove(file))

			continue
		}

		require.NoError(h.t, os.MkdirAll(filepath.Dir(file), 0o750))
		require.NoError(h.t, os.WriteFile(file, []byte(content), 0o600))
	}

	h.git("add", "--all")
	h.git("commit", "--quiet", "--allow-empty", "--message", message)

	return h.git("rev-parse", "HEAD")
}

type images map[string]bool

func findImageCommit(t *testing.T, h history, published images) (string, error) {
	t.Helper()

	return imageCommit(t.Context(), h.repo, func(tag string) (bool, error) { return published[tag], nil },
		"HEAD", testInputs)
}

func TestTheImageCommitIsTheFirstCommitWithAnImage(t *testing.T) {
	t.Parallel()

	// Verifies: RLS-0021
	t.Run("an image at the commit", func(t *testing.T) {
		t.Parallel()

		h := newHistory(t)
		h.commit("services/register/main.go")
		head := h.commit("go.mod")

		got, err := findImageCommit(t, h, images{head: true})

		require.NoError(t, err)
		assert.Equal(t, head, got)
	})

	// Verifies: RLS-0021, RLS-0022
	t.Run("an image at an earlier commit, with no change to an input after it", func(t *testing.T) {
		t.Parallel()

		h := newHistory(t)
		image := h.commit("services/register/main.go")
		h.commit("services/register/README.md")
		h.commit("charts/register/values.yaml")

		got, err := findImageCommit(t, h, images{image: true})

		require.NoError(t, err)
		assert.Equal(t, image, got)
	})

	// Verifies: RLS-0021, RLS-0022
	t.Run("an image of a commit that changes no input, from a build with no base", func(t *testing.T) {
		t.Parallel()

		h := newHistory(t)
		older := h.commit("services/register/main.go")
		rebuilt := h.commit("charts/register/values.yaml")

		got, err := findImageCommit(t, h, images{older: true, rebuilt: true})

		require.NoError(t, err)
		assert.Equal(t, rebuilt, got)
	})
}

// Verifies: RLS-0022, REG-1513, DH-2645
func TestAChangeToAnInputAfterTheImageCommitIsRefused(t *testing.T) {
	t.Parallel()

	h := newHistory(t)
	image := h.commit("services/register/main.go")
	changed := h.commit("go.mod")
	h.commit("docs/notes.txt")

	_, err := findImageCommit(t, h, images{image: true})

	require.Error(t, err)
	assert.Contains(t, err.Error(), changed)
}

func TestNoCommitThatChangesAnInputIsRefused(t *testing.T) {
	t.Parallel()

	h := newHistory(t)
	head := h.commit("charts/register/values.yaml")

	_, err := findImageCommit(t, h, images{head: true})

	require.Error(t, err)
}

// Verifies: RLS-0023
func TestTheImageInputsFileHasEachService(t *testing.T) {
	t.Parallel()

	inputs, err := readInputs("../../.github/image-inputs.json")
	require.NoError(t, err)

	entries, err := os.ReadDir("../../services")
	require.NoError(t, err)

	for _, e := range entries {
		if _, err := os.Stat(filepath.Join("../../services", e.Name(), "Dockerfile")); err != nil {
			continue
		}

		assert.Contains(t, inputs[e.Name()], "services/"+e.Name(), e.Name())
		assert.Contains(t, inputs[e.Name()], ".github/image-inputs.json", e.Name())
		assert.Contains(t, inputs[e.Name()], ".github/actions/publish-image", e.Name())
	}
}

func TestAnUnknownServiceIsRefused(t *testing.T) {
	t.Parallel()

	err := run(t.Context(), []string{
		"image-commit", "-service", "nothing", "-image", "example.test/x",
		"-inputs", "../../.github/image-inputs.json",
	}, &strings.Builder{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nothing")
}

func TestTheRegistryGivesAnImageOfATagWithAnAnonymousToken(t *testing.T) {
	t.Parallel()

	const token = "anonymous-token"

	var srv *httptest.Server

	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/token":
			assert.Equal(t, "registry.test", r.URL.Query().Get("service"))
			assert.Equal(t, "repository:team/register:pull", r.URL.Query().Get("scope"))

			_, _ = w.Write([]byte(`{"token":"` + token + `"}`))
		case r.Header.Get("Authorization") != "Bearer "+token:
			w.Header().Set("WWW-Authenticate",
				`Bearer realm="`+srv.URL+`/token",service="registry.test",scope="repository:team/register:pull"`)
			w.WriteHeader(http.StatusUnauthorized)
		case r.Method == http.MethodHead && r.URL.Path == "/v2/team/register/manifests/published":
			assert.Contains(t, r.Header.Get("Accept"), "application/vnd.oci.image.index.v1+json")
			w.Header().Set("Docker-Content-Digest", "sha256:0123")
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/v2/team/register/manifests/broken":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	reg, err := newRegistry(srv.Client(), strings.TrimPrefix(srv.URL, "https://")+"/team/register")
	require.NoError(t, err)

	for tag, want := range map[string]bool{"published": true, "missing": false} {
		got, err := reg.has(t.Context(), tag)

		require.NoError(t, err, tag)
		assert.Equal(t, want, got, tag)
	}

	_, err = reg.has(t.Context(), "broken")
	require.Error(t, err)

	digest, err := reg.digest(t.Context(), "published")
	require.NoError(t, err)
	assert.Equal(t, "sha256:0123", digest)

	_, err = reg.digest(t.Context(), "missing")
	require.Error(t, err)
}
