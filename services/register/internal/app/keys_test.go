package app_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

func keys(states map[string]app.KeyState) []app.Key {
	// The order of the JWKS: the key that signs, the next key, the previous key.
	list := []app.Key{}

	for _, state := range []app.KeyState{app.KeySigning, app.KeyNext, app.KeyPrevious} {
		for id, s := range states {
			if s == state {
				list = append(list, app.Key{ID: id, State: s})
			}
		}
	}

	return list
}

func TestAPublishRemovesThePreviousKey(t *testing.T) {
	changes, err := app.PlanKeyChange(keys(map[string]app.KeyState{"k2": app.KeySigning, "k1": app.KeyPrevious}),
		app.PublishNext)

	require.NoError(t, err)
	assert.Equal(t, []app.KeyChange{{KeyID: "k1", From: app.KeyPrevious}}, changes)
}

func TestAPublishWithNoPreviousKeyChangesNoKeyThatExists(t *testing.T) {
	changes, err := app.PlanKeyChange(keys(map[string]app.KeyState{"k1": app.KeySigning}), app.PublishNext)

	require.NoError(t, err)
	assert.Empty(t, changes)
}

func TestAnActivationChangesTheSigningKeyFirst(t *testing.T) {
	changes, err := app.PlanKeyChange(keys(map[string]app.KeyState{"k1": app.KeySigning, "k2": app.KeyNext}),
		app.ActivateNext)

	require.NoError(t, err)
	// A key set can have only one key that signs, so the order is the order of the change.
	assert.Equal(t, []app.KeyChange{
		{KeyID: "k1", From: app.KeySigning, To: app.KeyPrevious},
		{KeyID: "k2", From: app.KeyNext, To: app.KeySigning},
	}, changes)
}

func TestKeyChangeStepsThatTheRulesRefuse(t *testing.T) {
	_, inProgress := app.PlanKeyChange(keys(map[string]app.KeyState{"k1": app.KeySigning, "k2": app.KeyNext}),
		app.PublishNext)
	_, noNext := app.PlanKeyChange(keys(map[string]app.KeyState{"k1": app.KeySigning}), app.ActivateNext)

	require.ErrorIs(t, inProgress, app.ErrKeyChangeInProgress)
	require.ErrorIs(t, noNext, app.ErrNoNextKey)
}

func TestTheKeySetOffersOnlyTheStepThatItsStatePermits(t *testing.T) {
	tests := map[string]struct {
		keys map[string]app.KeyState
		want []app.KeyActionPlan
	}{
		"one key": {
			keys: map[string]app.KeyState{"k1": app.KeySigning},
			want: []app.KeyActionPlan{{Action: app.PublishNext, Changes: []app.KeyChange{}}},
		},
		"a next key": {
			keys: map[string]app.KeyState{"k1": app.KeySigning, "k2": app.KeyNext},
			want: []app.KeyActionPlan{{Action: app.ActivateNext, Changes: []app.KeyChange{
				{KeyID: "k1", From: app.KeySigning, To: app.KeyPrevious},
				{KeyID: "k2", From: app.KeyNext, To: app.KeySigning},
			}}},
		},
		"a previous key": {
			keys: map[string]app.KeyState{"k2": app.KeySigning, "k1": app.KeyPrevious},
			want: []app.KeyActionPlan{{Action: app.PublishNext, Changes: []app.KeyChange{
				{KeyID: "k1", From: app.KeyPrevious},
			}}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, app.PermittedKeyActions(keys(tt.keys)))
		})
	}
}
