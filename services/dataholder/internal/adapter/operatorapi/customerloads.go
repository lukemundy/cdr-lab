package operatorapi

import (
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/app"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

// A load is one request for all the seed customers of a brand, which a Kubernetes ConfigMap of 1 MiB holds.
const maxLoadBytes = 1 << 20

type loadResponse struct {
	Customers []app.LoadResult `json:"customers"`
}

func loadCustomers(customers Customers) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body domain.CustomerLoad
		if !decodeUpTo(w, r, &body, maxLoadBytes) {
			return
		}

		results, err := customers.Load(r.Context(), body)
		if err != nil {
			writeError(w, r, err)
			return
		}

		jsonhttp.Write(w, http.StatusOK, loadResponse{Customers: results})
	})
}
