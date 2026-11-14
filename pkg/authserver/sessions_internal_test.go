package authserver

import (
	"testing"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

func newTestSessions(limit int) (*sessionStore, *clock) {
	c := &clock{now: time.Date(2026, 10, 10, 9, 0, 0, 0, time.UTC)}

	return newSessionStore(limit, c.Now), c
}

func session(id string, expiresAt time.Time) *goidc.AuthnSession {
	return &goidc.AuthnSession{ID: id, PushedAuthReqID: "par-" + id, ExpiresAt: int(expiresAt.Unix())}
}

func TestASessionIsFoundByItsIDAndByItsPushedRequestID(t *testing.T) {
	sessions, c := newTestSessions(10)
	saved := session("a", c.now.Add(time.Minute))

	require.NoError(t, sessions.SaveSession(t.Context(), saved))

	byID, err := sessions.Session(t.Context(), "a")
	require.NoError(t, err)
	assert.Same(t, saved, byID)

	byRequest, err := sessions.SessionByPushedAuthReqID(t.Context(), "par-a")
	require.NoError(t, err)
	assert.Same(t, saved, byRequest)
}

func TestASessionThatDoesNotExistIsNotFound(t *testing.T) {
	sessions, _ := newTestSessions(10)

	_, byID := sessions.Session(t.Context(), "a")
	_, byRequest := sessions.SessionByPushedAuthReqID(t.Context(), "par-a")

	require.ErrorIs(t, byID, goidc.ErrNotFound)
	require.ErrorIs(t, byRequest, goidc.ErrNotFound)
}

func TestAnExpiredSessionIsNotFound(t *testing.T) {
	sessions, c := newTestSessions(10)
	require.NoError(t, sessions.SaveSession(t.Context(), session("a", c.now.Add(time.Minute))))

	c.now = c.now.Add(time.Minute)

	_, byID := sessions.Session(t.Context(), "a")
	_, byRequest := sessions.SessionByPushedAuthReqID(t.Context(), "par-a")

	require.ErrorIs(t, byID, goidc.ErrNotFound)
	require.ErrorIs(t, byRequest, goidc.ErrNotFound)
}

func TestAFullStoreRefusesANewSessionAndKeepsTheSessionsThatItHas(t *testing.T) {
	sessions, c := newTestSessions(2)
	require.NoError(t, sessions.SaveSession(t.Context(), session("a", c.now.Add(time.Minute))))
	require.NoError(t, sessions.SaveSession(t.Context(), session("b", c.now.Add(time.Minute))))

	err := sessions.SaveSession(t.Context(), session("c", c.now.Add(time.Minute)))

	require.ErrorIs(t, err, errSessionStoreFull)

	_, err = sessions.Session(t.Context(), "a")
	require.NoError(t, err)
	_, err = sessions.Session(t.Context(), "c")
	require.ErrorIs(t, err, goidc.ErrNotFound)
}

func TestAFullStoreAcceptsAChangeToASessionThatItHas(t *testing.T) {
	sessions, c := newTestSessions(1)
	require.NoError(t, sessions.SaveSession(t.Context(), session("a", c.now.Add(time.Minute))))

	changed := session("a", c.now.Add(2*time.Minute))

	require.NoError(t, sessions.SaveSession(t.Context(), changed))

	got, err := sessions.Session(t.Context(), "a")
	require.NoError(t, err)
	assert.Same(t, changed, got)
}

func TestAnExpiredSessionGivesItsPlaceToANewSession(t *testing.T) {
	sessions, c := newTestSessions(1)
	require.NoError(t, sessions.SaveSession(t.Context(), session("a", c.now.Add(time.Minute))))

	c.now = c.now.Add(time.Minute)

	require.NoError(t, sessions.SaveSession(t.Context(), session("b", c.now.Add(time.Minute))))
}

func TestASessionWithANewPushedRequestIDIsNotFoundByTheOldOne(t *testing.T) {
	sessions, c := newTestSessions(10)
	require.NoError(t, sessions.SaveSession(t.Context(), session("a", c.now.Add(time.Minute))))

	changed := session("a", c.now.Add(time.Minute))
	changed.PushedAuthReqID = ""
	require.NoError(t, sessions.SaveSession(t.Context(), changed))

	_, err := sessions.SessionByPushedAuthReqID(t.Context(), "par-a")

	require.ErrorIs(t, err, goidc.ErrNotFound)
}
