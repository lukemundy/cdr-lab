package console_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/authserver"
	"github.com/lukemundy/cdr-lab/pkg/platform/database"
	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/console"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/idp"
	"github.com/lukemundy/cdr-lab/services/register/internal/adapter/ssa"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
	"github.com/lukemundy/cdr-lab/services/register/internal/migrations"
)

type keyReader map[app.KeySet][]app.Key

func (r keyReader) KeySet(_ context.Context, set app.KeySet) (app.KeySetState, error) {
	return app.KeySetState{Version: 7, Keys: r[set]}, nil
}

func getKeySets(t *testing.T, keys app.KeyReader) response {
	t.Helper()

	return serve(t, nil, keys, "/api/key-sets")
}

func TestKeySetsGiveEachKeyWithItsStateAndCreationTime(t *testing.T) {
	created := time.Date(2026, 10, 23, 1, 2, 3, 456000000, time.UTC)
	keys := keyReader{
		app.AccessTokenKeys: {
			{ID: "at-2", State: app.KeySigning, CreatedAt: created.Add(time.Hour)},
			{ID: "at-1", State: app.KeyPrevious, CreatedAt: created},
		},
		app.SSAKeys: {
			{ID: "ssa-1", State: app.KeySigning, CreatedAt: created},
			{ID: "ssa-2", State: app.KeyNext, CreatedAt: created.Add(time.Minute)},
		},
	}

	resp := getKeySets(t, keys)

	assert.Equal(t, http.StatusOK, resp.status)
	assert.Equal(t, "no-store", resp.header.Get("Cache-Control"))
	assert.JSONEq(t, `{"data": [
		{"keySet": "access-token", "version": 7, "keys": [
			{"keyId": "at-2", "state": "signing", "createdAt": "2026-10-23T02:02:03.456Z"},
			{"keyId": "at-1", "state": "previous", "createdAt": "2026-10-23T01:02:03.456Z"}
		], "actions": [
			{"action": "publish-next", "irreversible": true, "changes": [{"keyId": "at-1", "from": "previous"}]}
		]},
		{"keySet": "ssa", "version": 7, "keys": [
			{"keyId": "ssa-1", "state": "signing", "createdAt": "2026-10-23T01:02:03.456Z"},
			{"keyId": "ssa-2", "state": "next", "createdAt": "2026-10-23T01:03:03.456Z"}
		], "actions": [
			{"action": "activate-next", "irreversible": true, "changes": [
				{"keyId": "ssa-1", "from": "signing", "to": "previous"},
				{"keyId": "ssa-2", "from": "next", "to": "signing"}
			]}
		]}
	]}`, resp.body)
}

func TestKeySetsGiveTheCreationTimeInUTC(t *testing.T) {
	perth := time.FixedZone("AWST", 8*60*60)
	keys := keyReader{
		app.AccessTokenKeys: {
			{ID: "at-1", State: app.KeySigning, CreatedAt: time.Date(2026, 10, 23, 9, 0, 0, 0, perth)},
		},
		app.SSAKeys: {
			{ID: "ssa-1", State: app.KeySigning, CreatedAt: time.Date(2026, 10, 23, 9, 0, 0, 0, perth)},
		},
	}

	resp := getKeySets(t, keys)

	assert.Contains(t, resp.body, `"createdAt":"2026-10-23T01:00:00Z"`)
}

type keySetList struct {
	Data []struct {
		KeySet app.KeySet `json:"keySet"`
		Keys   []struct {
			KeyID string       `json:"keyId"`
			State app.KeyState `json:"state"`
		} `json:"keys"`
	} `json:"data"`
}

// states gives the state of each key ID in each key set.
func states(t *testing.T, resp response) map[app.KeySet]map[string]app.KeyState {
	t.Helper()
	require.Equal(t, http.StatusOK, resp.status)

	var list keySetList
	require.NoError(t, json.Unmarshal([]byte(resp.body), &list))

	got := map[app.KeySet]map[string]app.KeyState{}
	for _, set := range list.Data {
		got[set.KeySet] = map[string]app.KeyState{}
		for _, k := range set.Keys {
			got[set.KeySet][k.KeyID] = k.State
		}
	}

	return got
}

func keyStore(t *testing.T) *idp.Store {
	t.Helper()

	pool := databasetest.New(t)
	require.NoError(t, database.Migrate(t.Context(), pool, migrations.FS))

	store := idp.NewStore(pool)
	for _, set := range []app.KeySet{app.AccessTokenKeys, app.SSAKeys} {
		require.NoError(t, store.EnsureSigningKey(t.Context(), set))
	}

	return store
}

// jwksKeyIDs gives the kid of each key in the JWKS of the set, which the JWKS endpoints serve.
func jwksKeyIDs(t *testing.T, store *idp.Store, set app.KeySet) []string {
	t.Helper()

	keys, err := store.SigningKeys(set)(t.Context())
	require.NoError(t, err)

	ids := []string{}
	for _, k := range authserver.PublicJWKS(keys).Keys {
		ids = append(ids, k.KeyID)
	}

	return ids
}

func onlyKeyOf(t *testing.T, keys map[string]app.KeyState) string {
	t.Helper()
	require.Len(t, keys, 1)

	for id := range keys {
		return id
	}

	return ""
}

func TestKeySetsFollowEachStepOfAKeyChange(t *testing.T) {
	for _, set := range []app.KeySet{app.AccessTokenKeys, app.SSAKeys} {
		t.Run(string(set), func(t *testing.T) {
			other := app.SSAKeys
			if set == app.SSAKeys {
				other = app.AccessTokenKeys
			}

			store := keyStore(t)
			before := states(t, getKeySets(t, store))
			first := onlyKeyOf(t, before[set])

			require.NoError(t, store.PublishNextKey(t.Context(), set))
			published := states(t, getKeySets(t, store))

			require.Len(t, published[set], 2)
			assert.Equal(t, app.KeySigning, published[set][first], "the first key still signs")

			var next string

			for id, state := range published[set] {
				if state == app.KeyNext {
					next = id
				}
			}

			require.NotEmpty(t, next, "the key set has a next key")
			assert.Contains(t, jwksKeyIDs(t, store, set), next, "the JWKS publishes the next key")
			assert.Equal(t, before[other], published[other], "the other key set does not change")

			require.NoError(t, store.ActivateNextKey(t.Context(), set))
			activated := states(t, getKeySets(t, store))

			assert.Equal(t, map[string]app.KeyState{next: app.KeySigning, first: app.KeyPrevious}, activated[set])
			assert.Equal(t, before[other], activated[other], "the other key set does not change")
		})
	}
}

func TestKeySetsGiveNoKeyMaterial(t *testing.T) {
	store := keyStore(t)
	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))

	resp := getKeySets(t, store)

	require.Equal(t, http.StatusOK, resp.status)

	for _, set := range []app.KeySet{app.AccessTokenKeys, app.SSAKeys} {
		keys, err := store.SigningKeys(set)(t.Context())
		require.NoError(t, err)
		require.NotEmpty(t, keys)

		for _, jwk := range authserver.PublicJWKS(keys).Keys {
			assert.NotContains(t, resp.body, jwk.N, "the response has no modulus")
		}

		for _, k := range keys {
			der, err := k.MarshalPKCS8()
			require.NoError(t, err)
			assert.NotContains(t, resp.body, base64.StdEncoding.EncodeToString(der), "the response has no private key")
			assert.NotContains(t, resp.body, base64.RawURLEncoding.EncodeToString(k.Key.D.Bytes()))
		}
	}

	for _, field := range []string{`"kty"`, `"n"`, `"e"`, `"d"`, `"kid"`} {
		assert.NotContains(t, resp.body, field, "the response has no JWK")
	}
}

type failingKeyReader struct{}

func (failingKeyReader) KeySet(context.Context, app.KeySet) (app.KeySetState, error) {
	return app.KeySetState{}, errors.New("a message from the database")
}

func TestKeySetsGiveAProblemWithNoErrorMessageWhenTheReadFails(t *testing.T) {
	resp := getKeySets(t, failingKeyReader{})

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.Equal(t, "application/problem+json", resp.header.Get("Content-Type"))
	assert.NotContains(t, resp.body, "database")
}

func keyWithState(keys map[string]app.KeyState, state app.KeyState) string {
	for id, s := range keys {
		if s == state {
			return id
		}
	}

	return ""
}

type keySetBody struct {
	KeySet  app.KeySet      `json:"keySet"`
	Version int64           `json:"version"`
	Actions json.RawMessage `json:"actions"`
}

func keySetsOf(t *testing.T, store *idp.Store) map[app.KeySet]keySetBody {
	t.Helper()

	resp := getKeySets(t, store)
	require.Equal(t, http.StatusOK, resp.status)

	var list struct {
		Data []keySetBody `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(resp.body), &list))

	sets := map[app.KeySet]keySetBody{}
	for _, s := range list.Data {
		sets[s.KeySet] = s
	}

	return sets
}

func postKeyAction(t *testing.T, store *idp.Store, set app.KeySet, action app.KeyAction, version int64) response {
	t.Helper()

	return post(t, console.Ports{Keys: store, KeyChanger: store},
		"/api/key-sets/"+string(set)+"/"+string(action), fmt.Sprintf(`{"version": %d}`, version))
}

func TestAKeySetGivesItsVersionAndTheOneStepThatItsStatePermits(t *testing.T) {
	store := keyStore(t)
	first := onlyKeyOf(t, states(t, getKeySets(t, store))[app.SSAKeys])

	before := keySetsOf(t, store)[app.SSAKeys]

	assert.Equal(t, int64(1), before.Version)
	assert.JSONEq(t, `[{"action": "publish-next", "irreversible": true, "changes": []}]`, string(before.Actions))

	require.NoError(t, store.PublishNextKey(t.Context(), app.SSAKeys))
	published := keySetsOf(t, store)[app.SSAKeys]
	next := keyWithState(states(t, getKeySets(t, store))[app.SSAKeys], app.KeyNext)

	assert.Equal(t, int64(2), published.Version)
	assert.JSONEq(t, fmt.Sprintf(`[{"action": "activate-next", "irreversible": true, "changes": [
		{"keyId": %q, "from": "signing", "to": "previous"},
		{"keyId": %q, "from": "next", "to": "signing"}
	]}]`, first, next), string(published.Actions))

	require.NoError(t, store.ActivateNextKey(t.Context(), app.SSAKeys))
	activated := keySetsOf(t, store)[app.SSAKeys]

	assert.Equal(t, int64(3), activated.Version)
	assert.JSONEq(t, fmt.Sprintf(`[{"action": "publish-next", "irreversible": true, "changes": [
		{"keyId": %q, "from": "previous"}
	]}]`, first), string(activated.Actions), "a publish names the previous key that leaves the JWKS")
}

func TestAFullKeyChangeInTheConsoleSignsANewSSAWithTheNewKey(t *testing.T) {
	store := keyStore(t)
	before := states(t, getKeySets(t, store))
	first := onlyKeyOf(t, before[app.SSAKeys])
	accessTokenKeys := jwksKeyIDs(t, store, app.AccessTokenKeys)

	publish := postKeyAction(t, store, app.SSAKeys, app.PublishNext, 1)

	require.Equal(t, http.StatusNoContent, publish.status, publish.body)
	assert.Empty(t, publish.body)
	assert.Equal(t, "no-store", publish.header.Get("Cache-Control"))

	published := states(t, getKeySets(t, store))[app.SSAKeys]
	require.Len(t, published, 2)

	next := keyWithState(published, app.KeyNext)
	require.NotEmpty(t, next)
	assert.ElementsMatch(t, []string{first, next}, jwksKeyIDs(t, store, app.SSAKeys), "the JWKS has the new key")
	assert.Equal(t, accessTokenKeys, jwksKeyIDs(t, store, app.AccessTokenKeys), "the other key set does not change")

	activate := postKeyAction(t, store, app.SSAKeys, app.ActivateNext, 2)

	require.Equal(t, http.StatusNoContent, activate.status, activate.body)
	assert.Equal(t, map[string]app.KeyState{next: app.KeySigning, first: app.KeyPrevious},
		states(t, getKeySets(t, store))[app.SSAKeys])

	signed, err := ssa.Signer{Keys: store.SigningKeys(app.SSAKeys)}.SignSSA(t.Context(), domain.SSAClaims{})
	require.NoError(t, err)
	parsed, err := jwt.ParseSigned(signed, []jose.SignatureAlgorithm{jose.PS256})
	require.NoError(t, err)
	assert.Equal(t, next, parsed.Headers[0].KeyID, "a new SSA has the kid of the new key")
	assert.Equal(t, accessTokenKeys, jwksKeyIDs(t, store, app.AccessTokenKeys), "the other key set does not change")
}

func TestAKeyActionWithAVersionThatIsNotCurrentChangesNothing(t *testing.T) {
	store := keyStore(t)
	require.Equal(t, http.StatusNoContent, postKeyAction(t, store, app.SSAKeys, app.PublishNext, 1).status)
	before := getKeySets(t, store).body

	for _, action := range app.KeyActions {
		resp := postKeyAction(t, store, app.SSAKeys, action, 1)

		require.Equal(t, http.StatusConflict, resp.status, action)
		assert.Contains(t, resp.body, "The key set changed after the version in the request.", action)
	}

	assert.Equal(t, before, getKeySets(t, store).body)
}

func TestAKeyActionThatTheRulesDoNotPermitGivesTheRule(t *testing.T) {
	store := keyStore(t)

	noNext := postKeyAction(t, store, app.SSAKeys, app.ActivateNext, 1)
	require.Equal(t, http.StatusNoContent, postKeyAction(t, store, app.SSAKeys, app.PublishNext, 1).status)
	inProgress := postKeyAction(t, store, app.SSAKeys, app.PublishNext, 2)

	require.Equal(t, http.StatusConflict, noNext.status)
	assert.Contains(t, noNext.body, "The key set has no next key.")
	require.Equal(t, http.StatusConflict, inProgress.status)
	assert.Contains(t, inProgress.body, "The key set has a next key that is not active.")
	assert.Equal(t, int64(2), keySetsOf(t, store)[app.SSAKeys].Version)
}

func TestAKeyActionForAKeySetThatDoesNotExistGivesNotFound(t *testing.T) {
	store := keyStore(t)

	for _, action := range app.KeyActions {
		resp := post(t, console.Ports{KeyChanger: store}, "/api/key-sets/other/"+string(action), `{"version": 1}`)

		assert.Equal(t, http.StatusNotFound, resp.status, action)
	}
}

func TestAKeyActionRefusesABodyThatIsNotValid(t *testing.T) {
	store := keyStore(t)

	for _, body := range []string{
		``, `{}`, `null`, `{"version": 0}`, `{"version": "1"}`, `{"version": 1, "keyId": "k"}`, `{"version": 1} {}`,
	} {
		resp := post(t, console.Ports{KeyChanger: store}, "/api/key-sets/ssa/publish-next", body)

		assert.Equal(t, http.StatusBadRequest, resp.status, "body %.40s", body)
		assert.Contains(t, resp.body, "the version of the key set")
	}

	assert.Equal(t, int64(1), keySetsOf(t, store)[app.SSAKeys].Version)
}

type failingKeyChanger struct{}

func (failingKeyChanger) ChangeKeyAt(context.Context, app.KeySet, app.KeyAction, int64) error {
	return errors.New("a message from the database")
}

func TestAKeyActionGivesAProblemWithNoErrorMessageWhenTheChangeFails(t *testing.T) {
	resp := post(t, console.Ports{KeyChanger: failingKeyChanger{}}, "/api/key-sets/ssa/publish-next", `{"version": 1}`)

	assert.Equal(t, http.StatusInternalServerError, resp.status)
	assert.NotContains(t, resp.body, "database")
}
