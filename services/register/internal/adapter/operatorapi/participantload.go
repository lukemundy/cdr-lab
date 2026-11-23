package operatorapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

type ParticipantLoader interface {
	LoadParticipants(ctx context.Context, p domain.Participants) (domain.ParticipantLoad, error)
}

type participantLoadResponse struct {
	Added    []string `json:"added"`
	Existing []string `json:"existing"`
}

func loadParticipants(loader ParticipantLoader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var file domain.Participants
		if !decodeBody(w, r, &file) {
			return
		}

		load, err := loader.LoadParticipants(r.Context(), file)
		// The errors of more than one legal entity can include a status rule. A record that is not valid alone gives
		// 400 also then, as for one create.
		if errors.Is(err, domain.ErrNotValid) {
			problem.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		writeResult(w, r, err, "participants", func(w http.ResponseWriter) {
			jsonhttp.Write(w, http.StatusOK, participantLoadResponse{
				Added:    nonNil(load.Added),
				Existing: nonNil(load.Existing),
			})
		})
	})
}

func nonNil(ids []string) []string {
	if ids == nil {
		return []string{}
	}

	return ids
}
