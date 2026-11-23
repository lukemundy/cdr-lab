// Package dataholdertest gives a fake of the customer load of the data holder operator interface, for tests. As the
// data holder does, it creates each customer of a load whose email address it does not have.
package dataholdertest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

const Token = "dataholder-operator-token"

type DataHolder struct {
	*httptest.Server

	mu      sync.Mutex
	emails  map[string]bool
	loads   [][]byte
	refusal int
	detail  string
}

func New(t *testing.T) *DataHolder {
	t.Helper()

	d := &DataHolder{emails: map[string]bool{}}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/customer-loads", d.load)
	d.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer "+Token {
			problem(w, http.StatusUnauthorized, "The request has no valid operator token.")
			return
		}

		mux.ServeHTTP(w, req)
	}))
	t.Cleanup(d.Close)

	return d
}

// Refuse makes each later load fail with the status and the detail.
func (d *DataHolder) Refuse(status int, detail string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.refusal, d.detail = status, detail
}

// Loads gives the body of each load request.
func (d *DataHolder) Loads() [][]byte {
	d.mu.Lock()
	defer d.mu.Unlock()

	return append([][]byte(nil), d.loads...)
}

func (d *DataHolder) load(w http.ResponseWriter, req *http.Request) {
	d.mu.Lock()
	defer d.mu.Unlock()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		problem(w, http.StatusBadRequest, "The request body is not valid JSON.")
		return
	}

	d.loads = append(d.loads, body)

	if d.refusal != 0 {
		problem(w, d.refusal, d.detail)
		return
	}

	var file struct {
		Customers []struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"customers"`
	}

	if json.Unmarshal(body, &file) != nil {
		problem(w, http.StatusBadRequest, "The request body is not valid JSON.")
		return
	}

	type result struct {
		CustomerNumber string `json:"customerNumber"`
		Created        bool   `json:"created"`
	}

	results := make([]result, 0, len(file.Customers))
	for _, c := range file.Customers {
		email := strings.ToLower(c.EmailAddress)
		results = append(
			results,
			result{CustomerNumber: fmt.Sprintf("%08d", len(d.emails)+1), Created: !d.emails[email]},
		)
		d.emails[email] = true
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]result{"customers": results})
}

func problem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"type": "about:blank", "title": http.StatusText(status), "status": status, "detail": detail,
	})
}
