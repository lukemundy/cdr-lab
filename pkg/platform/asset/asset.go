// Package asset serves a file that is fixed while the process runs, for example a stylesheet or a logo.
package asset

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// Handler sets no Content-Security-Policy when the policy is empty.
func Handler(contentType string, data []byte, policy string) http.Handler {
	sum := sha256.Sum256(data)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("Content-Type", contentType)
		header.Set("X-Content-Type-Options", "nosniff")
		// The browser asks again each time, so a new file shows after a restart. The ETag makes the answer small.
		header.Set("Cache-Control", "no-cache")
		header.Set("ETag", etag)

		if policy != "" {
			header.Set("Content-Security-Policy", policy)
		}

		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
	})
}
