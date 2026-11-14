package authserver

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/luikyv/go-oidc/pkg/goidc"
)

var errSessionStoreFull = errors.New("authserver: the session store has no place for a new authorisation session")

// sessionStore keeps authorisation sessions in memory, so it is for one instance and loses them at a restart. A full
// store refuses a new session, so that a client with many requests cannot remove the session of a different client.
type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]*goidc.AuthnSession
	limit    int
	now      func() time.Time
}

var _ goidc.PARManager = (*sessionStore)(nil)

func newSessionStore(limit int, now func() time.Time) *sessionStore {
	return &sessionStore{sessions: map[string]*goidc.AuthnSession{}, limit: limit, now: now}
}

func (s *sessionStore) expired(session *goidc.AuthnSession) bool {
	return s.now().Unix() >= int64(session.ExpiresAt)
}

func (s *sessionStore) SaveSession(_ context.Context, session *goidc.AuthnSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[session.ID]; !exists && s.full() {
		return errSessionStoreFull
	}

	s.sessions[session.ID] = session

	return nil
}

// full removes the expired sessions before it gives its answer.
func (s *sessionStore) full() bool {
	if len(s.sessions) < s.limit {
		return false
	}

	for id, session := range s.sessions {
		if s.expired(session) {
			delete(s.sessions, id)
		}
	}

	return len(s.sessions) >= s.limit
}

func (s *sessionStore) Session(_ context.Context, id string) (*goidc.AuthnSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok || s.expired(session) {
		return nil, goidc.ErrNotFound
	}

	return session, nil
}

func (s *sessionStore) SessionByPushedAuthReqID(_ context.Context, id string) (*goidc.AuthnSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, session := range s.sessions {
		if session.PushedAuthReqID == id && !s.expired(session) {
			return session, nil
		}
	}

	return nil, goidc.ErrNotFound
}
