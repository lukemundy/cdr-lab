// Package registertest gives a fake of the certificate paths of the Register operator interface, for tests. It
// issues with the CA library, as the Register does.
package registertest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/ca"
)

const Token = "operator-token"

type Register struct {
	*httptest.Server

	mu           sync.Mutex
	Root         ca.Authority
	Intermediate ca.Authority
	issues       []string
	lastCSR      []byte
	refusal      int
	detail       string
}

func New(t *testing.T) *Register {
	t.Helper()

	r := &Register{}
	r.NewCA(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/cdr-ca-bundle", r.bundle)
	mux.HandleFunc("POST /v1/data-holders/{legalEntityId}/brands/{brandId}/certificates/issue", r.issue(ca.Server))
	mux.HandleFunc(
		"POST /v1/data-recipients/{legalEntityId}/software-products/{softwareProductId}/certificates/issue",
		r.issue(ca.Client),
	)
	r.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer "+Token {
			problem(w, http.StatusUnauthorized, "The request has no valid operator token.")
			return
		}

		mux.ServeHTTP(w, req)
	}))
	t.Cleanup(r.Close)

	return r
}

// NewCA replaces the CA of the fake, as a new deployment of the CA does.
func (r *Register) NewCA(t *testing.T) {
	t.Helper()

	// RSA 2048 keys keep the tests fast.
	var keys [2]*rsa.PrivateKey
	for i := range keys {
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		keys[i] = k
	}

	root, intermediate, err := ca.NewCDR(keys[0], keys[1], time.Now().UTC().Truncate(time.Second).Add(-time.Hour))
	require.NoError(t, err)

	intermediate.CRLURL = "http://register.test/crl/cdr-intermediate.crl"

	r.mu.Lock()
	defer r.mu.Unlock()

	r.Root, r.Intermediate = root, intermediate
}

// Refuse makes each later issue fail with the status and the detail.
func (r *Register) Refuse(status int, detail string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refusal, r.detail = status, detail
}

// Issues gives the method and the path of each issue request.
func (r *Register) Issues() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.issues...)
}

func (r *Register) LastCSR() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.lastCSR
}

func (r *Register) bundle(w http.ResponseWriter, _ *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()

	chain(w, ca.EncodeCertificates(r.Intermediate.Certificate, r.Root.Certificate))
}

func (r *Register) issue(kind ca.Kind) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		r.mu.Lock()
		defer r.mu.Unlock()

		r.issues = append(r.issues, req.Method+" "+req.URL.Path)

		if r.refusal != 0 {
			problem(w, r.refusal, r.detail)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			problem(w, http.StatusBadRequest, "The request body must be one PEM CSR.")
			return
		}

		r.lastCSR = body

		csr, err := ca.ParseCSR(body)
		if err != nil {
			problem(w, http.StatusBadRequest, "The request body must be one PEM CSR.")
			return
		}

		cert, err := r.Intermediate.Issue(csr, kind, time.Now().UTC().Truncate(time.Second))
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}

		chain(w, ca.EncodeCertificates(cert, r.Intermediate.Certificate))
	}
}

func chain(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/pem-certificate-chain")
	_, _ = w.Write(data)
}

func problem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type": "about:blank", "title": http.StatusText(status), "status": status, "detail": detail,
	})
}
