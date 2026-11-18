// Package operatorhttp has the HTTP rules that the operator interfaces of the services share.
package operatorhttp

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/platform/hostallowlist"
	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
)

const (
	TokenEnv = "OPERATOR_TOKEN"
	// AllowedHostsEnv is the allowlist of the internal listener, which the operator console also uses.
	AllowedHostsEnv = "INTERNAL_ALLOWED_HOSTS"
)

type Config struct {
	Token        string
	AllowedHosts []string
}

func LoadConfig(allowedHostsEnv string) (Config, error) {
	token := os.Getenv(TokenEnv)
	if token == "" {
		return Config{}, fmt.Errorf("set %s to the bearer token of the operator interface", TokenEnv)
	}

	hosts, err := hostallowlist.Parse(os.Getenv(allowedHostsEnv))
	if err != nil {
		return Config{}, fmt.Errorf(
			"set %s to the Host header values that clients send to the operator interface, separated by commas, "+
				"for example a host and a port, or a host with no port behind a TLS proxy: %w",
			allowedHostsEnv,
			err,
		)
	}

	return Config{Token: token, AllowedHosts: hosts}, nil
}

func Guard(cfg Config) httproute.Middleware {
	allowHost, checkToken := hostallowlist.Middleware(cfg.AllowedHosts), requireToken(cfg.Token)

	return func(next http.Handler) http.Handler {
		return allowHost(checkToken(next))
	}
}

func requireToken(token string) httproute.Middleware {
	// Hashes have the same length, so that the comparison time does not show the length of the token.
	want := sha256.Sum256([]byte(token))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// RFC 6750, section 3.1: a request with no bearer token gets a challenge with no error code.
			scheme, credentials, _ := strings.Cut(r.Header.Get("Authorization"), " ")
			if !strings.EqualFold(scheme, "Bearer") {
				w.Header().Set("WWW-Authenticate", "Bearer")
				problem.Write(w, http.StatusUnauthorized, "The request must have the operator token as a bearer token.")

				return
			}

			got := sha256.Sum256([]byte(credentials))
			if subtle.ConstantTimeCompare(got[:], want[:]) != 1 {
				w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
				problem.Write(w, http.StatusUnauthorized, "The bearer token is not the operator token.")

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
