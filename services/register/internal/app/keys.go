package app

import (
	"context"
	"errors"
	"time"
)

// KeySet is a set of signing keys that the Register publishes in one JWKS.
type KeySet string

const (
	AccessTokenKeys KeySet = "access-token"
	SSAKeys         KeySet = "ssa"
)

func ParseKeySet(s string) (KeySet, bool) {
	set := KeySet(s)
	return set, set == AccessTokenKeys || set == SSAKeys
}

var (
	ErrKeyChangeInProgress = errors.New("the key set has a next key that is not active")
	ErrNoNextKey           = errors.New("the key set has no next key")
	ErrKeySetChanged       = errors.New("the key set is not at the expected version")
)

// KeyAction is a step of a key change, with the same name as in the paths of the operator interface.
type KeyAction string

const (
	PublishNext  KeyAction = "publish-next"
	ActivateNext KeyAction = "activate-next"
)

var KeyActions = []KeyAction{PublishNext, ActivateNext}

// KeyChanger changes the signing key of a key set in two steps, so that each party
// sees the new key in the JWKS before the Register signs with it. Each step changes the version of the key set.
type KeyChanger interface {
	// PublishNextKey gives ErrKeyChangeInProgress when the key set has a next key.
	PublishNextKey(ctx context.Context, set KeySet) error
	// ActivateNextKey gives ErrNoNextKey when the key set has no next key.
	ActivateNextKey(ctx context.Context, set KeySet) error
}

// KeyChange is a change of a key that exists. To is empty when the key leaves the JWKS and the Register deletes it.
type KeyChange struct {
	KeyID string
	From  KeyState
	To    KeyState
}

type KeyActionPlan struct {
	Action  KeyAction
	Changes []KeyChange
}

// PlanKeyChange gives the changes of the keys in the sequence in which they must occur. A publish also adds a new
// next key, which is not in the changes because it does not exist yet.
func PlanKeyChange(keys []Key, action KeyAction) ([]KeyChange, error) {
	changes := []KeyChange{}

	switch action {
	case PublishNext:
		if hasState(keys, KeyNext) {
			return nil, ErrKeyChangeInProgress
		}

		for _, k := range keys {
			if k.State == KeyPrevious {
				changes = append(changes, KeyChange{KeyID: k.ID, From: KeyPrevious})
			}
		}
	case ActivateNext:
		if !hasState(keys, KeyNext) {
			return nil, ErrNoNextKey
		}
		// A key set can have only one key that signs, so the key that signs changes first.
		for _, from := range []KeyState{KeySigning, KeyNext} {
			for _, k := range keys {
				if k.State == from {
					changes = append(changes, KeyChange{KeyID: k.ID, From: from, To: stateAfterActivation[from]})
				}
			}
		}
	}

	return changes, nil
}

var stateAfterActivation = map[KeyState]KeyState{KeySigning: KeyPrevious, KeyNext: KeySigning}

func PermittedKeyActions(keys []Key) []KeyActionPlan {
	permitted := []KeyActionPlan{}

	for _, action := range KeyActions {
		if changes, err := PlanKeyChange(keys, action); err == nil {
			permitted = append(permitted, KeyActionPlan{Action: action, Changes: changes})
		}
	}

	return permitted
}

func hasState(keys []Key, state KeyState) bool {
	for _, k := range keys {
		if k.State == state {
			return true
		}
	}

	return false
}

// KeyState is the role of a key in its key set. The JWKS publishes the keys in each state.
type KeyState string

const (
	KeySigning KeyState = "signing"
	KeyNext    KeyState = "next"
	// KeyPrevious is the key that signed before the last activation.
	KeyPrevious KeyState = "previous"
)

// Key has no key material.
type Key struct {
	ID        string
	State     KeyState
	CreatedAt time.Time
}

// KeySetState has the keys with the key that signs first, then the next key, then the previous key.
type KeySetState struct {
	Version int64
	Keys    []Key
}

type KeyReader interface {
	KeySet(ctx context.Context, set KeySet) (KeySetState, error)
}
