// Package consolehttp serves a browser UI, for example an operator console: a web UI and its same-origin API as one
// group of routes on a listener, with no credential.
package consolehttp

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/hostallowlist"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
)

const (
	contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; " +
		"font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'; " +
		"frame-ancestors 'none'; require-trusted-types-for 'script'"
	indexFile = "index.html"
	// The UI build puts each file with a content hash in its name in this directory, and only those files.
	hashedDir = "assets/"
)

// Settings name the settings of one kind of UI, and the UI in the errors.
type Settings struct {
	// Name is the UI in the errors, for example "console".
	Name string
	// EnabledEnv turns on the UI, because a production service has no such UI.
	EnabledEnv string
	// AllowedHostsEnv has no default, because a default is a value of one deployment.
	AllowedHostsEnv string
	// ExampleHost is a Host header value in the error of a missing allowlist.
	ExampleHost string
}

var Console = Settings{
	Name:       "console",
	EnabledEnv: "CONSOLE_ENABLED",
	// The console shares the allowlist of the internal listener with the operator interface.
	AllowedHostsEnv: "INTERNAL_ALLOWED_HOSTS",
	ExampleHost:     "127.0.0.1:8080",
}

type Config struct {
	Enabled bool
	// AllowedHosts are Host header values in lower case.
	AllowedHosts []string
}

// LoadConfig needs AllowedHostsEnv only when the UI is on.
func (s Settings) LoadConfig() (Config, error) {
	v := os.Getenv(s.EnabledEnv)
	if v == "" {
		return Config{}, nil
	}

	on, err := strconv.ParseBool(v)
	if err != nil {
		return Config{}, fmt.Errorf("%s: %w", s.EnabledEnv, err)
	}

	if !on {
		return Config{}, nil
	}

	hosts, err := s.LoadAllowedHosts()
	if err != nil {
		return Config{}, err
	}

	return Config{Enabled: true, AllowedHosts: hosts}, nil
}

// LoadAllowedHosts reads only the allowlist, for a UI that has no EnabledEnv because it is always on.
func (s Settings) LoadAllowedHosts() ([]string, error) {
	hosts, err := hostallowlist.Parse(os.Getenv(s.AllowedHostsEnv))
	if err != nil {
		return nil, fmt.Errorf(
			"set %s to the Host header values that browsers send to the %s, separated by commas, "+
				"for example %s, or a host with no port behind a TLS proxy: %w",
			s.AllowedHostsEnv,
			s.Name,
			s.ExampleHost,
			err,
		)
	}

	return hosts, nil
}

// Register refuses a request whose Host is not in allowedHosts, so that DNS rebinding cannot reach the UI.
func (s Settings) Register(mux *http.ServeMux, ui fs.FS, allowedHosts []string, api func(*httproute.Routes)) error {
	if len(allowedHosts) == 0 {
		return fmt.Errorf("the %s has no allowed hosts; set %s", s.Name, s.AllowedHostsEnv)
	}

	index, err := fs.ReadFile(ui, indexFile)
	if err != nil {
		return fmt.Errorf("the %s UI has no %s; build the UI before the binary: %w", s.Name, indexFile, err)
	}

	routes := httproute.New(mux)
	routes.Use(securityHeaders, hostallowlist.Middleware(allowedHosts), refuseCrossOrigin(s.Name))

	apiRoutes := routes.With(cacheControl("no-store"))
	api(apiRoutes)
	apiRoutes.Handle("/api/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		problem.Write(w, http.StatusNotFound, "The URL and method are not a "+s.Name+" API operation.")
	}))

	files := serveUI(ui, index)

	// One pattern for all methods: a "GET /" pattern conflicts with "/api/" in http.ServeMux.
	routes.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			problem.Write(w, http.StatusMethodNotAllowed, "The UI accepts only GET and HEAD.")

			return
		}

		files.ServeHTTP(w, r)
	}))

	return nil
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func refuseCrossOrigin(name string) httproute.Middleware {
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		problem.Write(w, http.StatusForbidden, "The "+name+" refuses a change from a different origin.")
	}))

	return protection.Handler
}

func cacheControl(value string) httproute.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", value)
			next.ServeHTTP(w, r)
		})
	}
}

// serveUI gives the HTML page for each path that is not a file, so that the UI routes in the browser.
func serveUI(ui fs.FS, index []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

		data, err := readFile(ui, name)
		if err != nil {
			name, data = indexFile, index
		}

		if strings.HasPrefix(name, hashedDir) {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	})
}

var errNotAFile = errors.New("not a file")

func readFile(ui fs.FS, name string) ([]byte, error) {
	info, err := fs.Stat(ui, name)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, errNotAFile
	}

	return fs.ReadFile(ui, name)
}
