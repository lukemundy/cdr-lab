package console

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lukemundy/cdr-lab/pkg/platform/httproute"
	"github.com/lukemundy/cdr-lab/pkg/platform/jsonhttp"
	"github.com/lukemundy/cdr-lab/pkg/platform/problem"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

type KeyChanger interface {
	// ChangeKeyAt returns app.ErrKeySetChanged when the key set is not at version.
	ChangeKeyAt(ctx context.Context, set app.KeySet, action app.KeyAction, version int64) error
}

const (
	noKeySet      = "The key set does not exist. The key sets are access-token and ssa."
	staleKeySet   = "The key set changed after the version in the request. The Register changed nothing."
	badKeySetBody = "The request body must be a JSON object with only the version of the key set, " +
		"an integer of 1 or more."
	keyChangeActive = "The key set has a next key that is not active. Activate it before you publish a new key."
	noNextKey       = "The key set has no next key. Publish a key before you activate it."
)

var keySetOrder = []app.KeySet{app.AccessTokenKeys, app.SSAKeys}

type keySetList struct {
	Data []keySet `json:"data"`
}

type keySet struct {
	KeySet  app.KeySet  `json:"keySet"`
	Version int64       `json:"version"`
	Keys    []key       `json:"keys"`
	Actions []keyAction `json:"actions"`
}

// key must have no key material: the JWKS endpoints give the public keys.
type key struct {
	KeyID     string       `json:"keyId"`
	State     app.KeyState `json:"state"`
	CreatedAt time.Time    `json:"createdAt"`
}

// Each key step is irreversible: an activation changes the key that participants must trust, and a publish
// deletes the previous key.
type keyAction struct {
	Action       app.KeyAction `json:"action"`
	Irreversible bool          `json:"irreversible"`
	Changes      []keyChange   `json:"changes"`
}

type keyChange struct {
	KeyID string       `json:"keyId"`
	From  app.KeyState `json:"from"`
	To    app.KeyState `json:"to,omitempty"`
}

func keySets(reader app.KeyReader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		list := keySetList{Data: make([]keySet, 0, len(keySetOrder))}

		for _, set := range keySetOrder {
			state, err := reader.KeySet(r.Context(), set)
			if err != nil {
				problem.WriteUnexpected(w, r, err, notCompleted)
				return
			}

			list.Data = append(list.Data, keySetOf(set, state))
		}

		jsonhttp.Write(w, http.StatusOK, list)
	})
}

func keySetOf(set app.KeySet, state app.KeySetState) keySet {
	item := keySet{
		KeySet:  set,
		Version: state.Version,
		Keys:    make([]key, 0, len(state.Keys)),
		Actions: []keyAction{},
	}
	for _, k := range state.Keys {
		item.Keys = append(item.Keys, key{KeyID: k.ID, State: k.State, CreatedAt: k.CreatedAt.UTC()})
	}

	for _, plan := range app.PermittedKeyActions(state.Keys) {
		action := keyAction{Action: plan.Action, Irreversible: true, Changes: make([]keyChange, 0, len(plan.Changes))}
		for _, c := range plan.Changes {
			action.Changes = append(action.Changes, keyChange{KeyID: c.KeyID, From: c.From, To: c.To})
		}

		item.Actions = append(item.Actions, action)
	}

	return item
}

func registerKeyActions(routes *httproute.Routes, changer KeyChanger) {
	for _, action := range app.KeyActions {
		routes.Handle("POST /api/key-sets/{keySet}/"+string(action), changeKey(changer, action))
	}
}

func changeKey(changer KeyChanger, action app.KeyAction) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		set, ok := app.ParseKeySet(r.PathValue("keySet"))
		if !ok {
			problem.Write(w, http.StatusNotFound, noKeySet)
			return
		}

		version, err := decodeVersion(http.MaxBytesReader(w, r.Body, maxActionBodyBytes))
		if err != nil {
			problem.WriteBodyRefusal(w, err, badKeySetBody)
			return
		}

		err = changer.ChangeKeyAt(r.Context(), set, action, version)

		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, app.ErrKeySetChanged):
			problem.Write(w, http.StatusConflict, staleKeySet)
		case errors.Is(err, app.ErrKeyChangeInProgress):
			problem.Write(w, http.StatusConflict, keyChangeActive)
		case errors.Is(err, app.ErrNoNextKey):
			problem.Write(w, http.StatusConflict, noNextKey)
		default:
			problem.WriteUnexpected(w, r, err, notCompleted)
		}
	})
}
