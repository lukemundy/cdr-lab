package banksite

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

const (
	maxSessions     = 10_000
	sessionLifetime = 24 * time.Hour
	sessionIDBytes  = 32
)

var errNoPlace = errors.New("the online bank site has the maximum number of sessions")

type session struct {
	customerNumber string
	endsAt         time.Time
}

// sessionStore keeps the sessions in memory, so it is for one instance and loses them at a restart. A full store
// refuses a new session, so that many sign-ins cannot end the sessions of other consumers.
type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]session
	// reserved is the number of places that a sign-in or a sign-up holds before its session starts.
	reserved int
	limit    int
	now      func() time.Time
}

func newSessionStore(limit int, now func() time.Time) *sessionStore {
	return &sessionStore{sessions: map[string]session{}, limit: limit, now: now}
}

// place is a reserved place in the store. A sign-up reserves it before it creates the customer, so that a full store
// creates no customer.
type place struct {
	store *sessionStore
	once  sync.Once
}

// reserve gives errNoPlace when the store is full.
func (s *sessionStore) reserve() (*place, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.full() {
		return nil, errNoPlace
	}

	s.reserved++

	return &place{store: s}, nil
}

// full removes the sessions that ended before it gives its answer.
func (s *sessionStore) full() bool {
	if len(s.sessions)+s.reserved < s.limit {
		return false
	}

	now := s.now()
	for id, session := range s.sessions {
		if !now.Before(session.endsAt) {
			delete(s.sessions, id)
		}
	}

	return len(s.sessions)+s.reserved >= s.limit
}

// start gives the ID of a new session in the place. The place is then used.
func (p *place) start(customerNumber string) string {
	id := newSessionID()

	p.once.Do(func() {
		s := p.store
		s.mu.Lock()
		defer s.mu.Unlock()

		s.reserved--
		s.sessions[id] = session{customerNumber: customerNumber, endsAt: s.now().Add(sessionLifetime)}
	})

	return id
}

// release gives the place back when no session started in it. A release after start does nothing.
func (p *place) release() {
	p.once.Do(func() {
		p.store.mu.Lock()
		defer p.store.mu.Unlock()

		p.store.reserved--
	})
}

func (s *sessionStore) customer(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[id]
	if !ok || !s.now().Before(session.endsAt) {
		return "", false
	}

	return session.customerNumber, true
}

func (s *sessionStore) end(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
}

func newSessionID() string {
	// Read gives no error: the Go runtime stops the process when the system has no random source.
	raw := make([]byte, sessionIDBytes)
	_, _ = rand.Read(raw)

	return base64.RawURLEncoding.EncodeToString(raw)
}
