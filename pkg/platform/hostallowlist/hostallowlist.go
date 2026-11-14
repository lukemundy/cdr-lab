// Package hostallowlist refuses a request whose Host header is not in an allowlist, so that DNS rebinding fails.
package hostallowlist

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
)

// Parse gives the values in lower case. A client sends a host with no port for the default port of its scheme.
func Parse(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, errors.New("the value is empty")
	}

	var hosts []string

	for entry := range strings.SplitSeq(value, ",") {
		entry = strings.ToLower(strings.TrimSpace(entry))
		if err := validate(entry); err != nil {
			return nil, err
		}

		hosts = append(hosts, entry)
	}

	return hosts, nil
}

func validate(entry string) error {
	h, port, err := net.SplitHostPort(entry)
	if err != nil {
		h, _, err = net.SplitHostPort(entry + ":0")
		if err != nil || h == "" {
			return fmt.Errorf("%q is not a host, or a host and a port", entry)
		}

		return nil
	}

	if h == "" || port == "" {
		return fmt.Errorf("%q is not a host, or a host and a port", entry)
	}

	if _, err := strconv.ParseUint(port, 10, 16); err != nil {
		return fmt.Errorf("%q does not have a valid port", entry)
	}

	return nil
}

func Middleware(hosts []string) httproute.Middleware {
	allowed := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		allowed[strings.ToLower(h)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowed[strings.ToLower(r.Host)] {
				problem.Write(
					w,
					http.StatusForbidden,
					"The server refuses a request for a host name that it does not serve.",
				)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
