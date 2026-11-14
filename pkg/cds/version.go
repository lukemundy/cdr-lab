package cds

import (
	"math"
	"net/http"
	"strconv"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
)

// Versioned is for endpoints that serve one representation for all their supported versions.
func Versioned(minV, maxV int) httproute.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			xv := r.Header.Get("x-v")
			if xv == "" {
				WriteError(w, http.StatusBadRequest, HeaderMissing, "x-v")
				return
			}

			high, ok := PositiveInteger(xv)
			if !ok {
				WriteError(w, http.StatusBadRequest, InvalidVersion, "x-v")
				return
			}

			low := high

			if xMinV := r.Header.Get("x-min-v"); xMinV != "" {
				v, ok := PositiveInteger(xMinV)
				if !ok {
					WriteError(w, http.StatusBadRequest, InvalidVersion, "x-min-v")
					return
				}
				// An x-min-v equal to or higher than x-v is ignored.
				low = min(v, high)
			}

			served := min(high, maxV)
			if served < low || served < minV {
				WriteError(
					w,
					http.StatusNotAcceptable,
					UnsupportedVersion,
					"The supported versions are "+strconv.Itoa(minV)+" to "+strconv.Itoa(maxV)+".",
				)

				return
			}

			w.Header().Set("x-v", strconv.Itoa(served))
			next.ServeHTTP(w, r)
		})
	}
}

// PositiveInteger gives math.MaxInt for a value that is too large for an int, which is still a positive integer.
func PositiveInteger(s string) (int, bool) {
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
	}

	v, err := strconv.Atoi(s)
	if err != nil {
		return math.MaxInt, true
	}

	return v, v > 0
}
