package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

func unexpected(w http.ResponseWriter, r *http.Request, err error) {
	cds.WriteUnexpected(w, r, err, "The Register could not complete the request.")
}

// no-cache makes clients revalidate the ETag each time, so that they see a status change at once.
func writeCacheable(w http.ResponseWriter, r *http.Request, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	sum := sha256.Sum256(data)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")

	if matchesAny(r.Header.Values("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// matchesAny applies the If-None-Match weak comparison of RFC 9110, section 13.1.2.
func matchesAny(ifNoneMatch []string, etag string) bool {
	for _, field := range ifNoneMatch {
		for tag := range strings.SplitSeq(field, ",") {
			tag = strings.TrimSpace(tag)
			if tag == "*" || strings.TrimPrefix(tag, "W/") == etag {
				return true
			}
		}
	}

	return false
}
