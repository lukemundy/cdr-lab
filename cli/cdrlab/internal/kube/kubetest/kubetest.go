// Package kubetest gives a fake of the Secrets API of Kubernetes for one namespace, for tests.
package kubetest

import (
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	Namespace = "lab"
	Token     = "service-account-token"
)

type Secret struct {
	Type            string
	Data            map[string][]byte
	ResourceVersion string
}

type object struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name            string `json:"name"`
		Namespace       string `json:"namespace,omitempty"`
		ResourceVersion string `json:"resourceVersion,omitempty"`
	} `json:"metadata"`
	Type string            `json:"type"`
	Data map[string][]byte `json:"data"`
}

// API is a fake of the Secrets of Namespace. It refuses a request that has no bearer Token.
type API struct {
	*httptest.Server

	// TokenFile has Token, as the projected token of a ServiceAccount has.
	TokenFile string

	mu      sync.Mutex
	secrets map[string]Secret
	version int
	refuse  map[string]int
	calls   []string
}

func New(t *testing.T) *API {
	t.Helper()

	return newAPI(t, httptest.NewServer)
}

// NewTLS gives the fake with TLS, as the API server of a cluster has.
func NewTLS(t *testing.T) *API {
	t.Helper()

	return newAPI(t, httptest.NewTLSServer)
}

func newAPI(t *testing.T, start func(http.Handler) *httptest.Server) *API {
	t.Helper()

	api := &API{secrets: map[string]Secret{}, refuse: map[string]int{}}
	api.TokenFile = filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(api.TokenFile, []byte(Token+"\n"), 0o600))

	mux := http.NewServeMux()
	path := "/api/v1/namespaces/" + Namespace + "/secrets"
	mux.HandleFunc("GET "+path+"/{name}", api.get)
	mux.HandleFunc("POST "+path, api.create)
	mux.HandleFunc("PUT "+path+"/{name}", api.update)
	mux.HandleFunc("DELETE "+path+"/{name}", api.remove)
	api.Server = start(api.authorise(mux))
	t.Cleanup(api.Close)

	return api
}

// CAPEM gives the CA certificate of a fake from NewTLS.
func (a *API) CAPEM() []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: a.Certificate().Raw})
}

func (a *API) Host() string {
	host, _, _ := net.SplitHostPort(a.Listener.Addr().String())
	return host
}

func (a *API) Port() string {
	_, port, _ := net.SplitHostPort(a.Listener.Addr().String())
	return port
}

// Refuse makes each later request with the method for the Secret fail with the status.
func (a *API) Refuse(method, name string, status int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.refuse[method+" "+name] = status
}

func (a *API) Put(name string, s Secret) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.version++
	s.ResourceVersion = strconv.Itoa(a.version)
	a.secrets[name] = s
}

func (a *API) Secret(name string) (Secret, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	s, ok := a.secrets[name]

	return s, ok
}

func (a *API) Names() []string {
	a.mu.Lock()
	defer a.mu.Unlock()

	names := make([]string, 0, len(a.secrets))
	for name := range a.secrets {
		names = append(names, name)
	}

	return names
}

// Calls gives the method and the Secret name of each request, for example "PUT tls".
func (a *API) Calls() []string {
	a.mu.Lock()
	defer a.mu.Unlock()

	return append([]string(nil), a.calls...)
}

func (a *API) authorise(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+Token {
			status(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *API) refused(w http.ResponseWriter, method, name string) bool {
	a.calls = append(a.calls, method+" "+name)

	code, ok := a.refuse[method+" "+name]
	if ok {
		status(w, code, fmt.Sprintf("refused %s of secrets %q", method, name))
	}

	return ok
}

func (a *API) get(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	name := r.PathValue("name")
	if a.refused(w, r.Method, name) {
		return
	}

	s, ok := a.secrets[name]
	if !ok {
		status(w, http.StatusNotFound, fmt.Sprintf("secrets %q not found", name))
		return
	}

	write(w, http.StatusOK, name, s)
}

func (a *API) create(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var o object
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil || o.Kind != "Secret" || o.APIVersion != "v1" {
		status(w, http.StatusBadRequest, "not a v1 Secret")
		return
	}

	name := o.Metadata.Name
	if a.refused(w, r.Method, name) {
		return
	}

	if _, ok := a.secrets[name]; ok {
		status(w, http.StatusConflict, fmt.Sprintf("secrets %q already exists", name))
		return
	}

	a.version++
	s := Secret{Type: o.Type, Data: o.Data, ResourceVersion: strconv.Itoa(a.version)}
	a.secrets[name] = s
	write(w, http.StatusCreated, name, s)
}

func (a *API) update(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	name := r.PathValue("name")
	if a.refused(w, r.Method, name) {
		return
	}

	var o object
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil || o.Metadata.Name != name {
		status(w, http.StatusBadRequest, "not the Secret of the URL")
		return
	}

	current, ok := a.secrets[name]

	switch {
	case !ok:
		status(w, http.StatusNotFound, fmt.Sprintf("secrets %q not found", name))
	case o.Metadata.ResourceVersion != current.ResourceVersion:
		status(w, http.StatusConflict, "the object has been modified")
	case o.Type != current.Type:
		status(w, http.StatusUnprocessableEntity, "field is immutable: type")
	default:
		a.version++
		s := Secret{Type: o.Type, Data: o.Data, ResourceVersion: strconv.Itoa(a.version)}
		a.secrets[name] = s
		write(w, http.StatusOK, name, s)
	}
}

func (a *API) remove(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	defer a.mu.Unlock()

	name := r.PathValue("name")
	if a.refused(w, r.Method, name) {
		return
	}

	if _, ok := a.secrets[name]; !ok {
		status(w, http.StatusNotFound, fmt.Sprintf("secrets %q not found", name))
		return
	}

	delete(a.secrets, name)
	status(w, http.StatusOK, "Success")
}

func write(w http.ResponseWriter, code int, name string, s Secret) {
	o := object{APIVersion: "v1", Kind: "Secret", Type: s.Type, Data: s.Data}
	o.Metadata.Name, o.Metadata.Namespace, o.Metadata.ResourceVersion = name, Namespace, s.ResourceVersion

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(o)
}

func status(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"apiVersion": "v1", "kind": "Status", "status": "Failure", "message": message, "code": code,
	})
}
