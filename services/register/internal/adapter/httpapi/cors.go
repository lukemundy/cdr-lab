package httpapi

import (
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/cds"
)

func AllowAnyOrigin(next http.Handler) http.Handler {
	return cds.AllowAnyOrigin(next, "x-v", "etag", "x-fapi-interaction-id")
}
