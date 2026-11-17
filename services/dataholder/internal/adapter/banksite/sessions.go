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
	// taken has each session whose place a reservation of a full store holds.
	taken map[string]bool
	limit int
	now   func() time.Time
}

func newSessionStore(limit int, now func() time.Time) *sessionStore {
	return &sessionStore{sessions: map[string]session{}, taken: map[string]bool{}, limit: limit, now: now}
}

// place is a reserved place in the store. A sign-up reserves it before it creates the customer, so that a full store
// creates no customer.
type place struct {
	store *sessionStore
	// replaces is the session of the request, which ends when the new session starts.
	replaces string
	once     sync.Once
}

// reserve gives errNoPlace when the store is full. The valid session current ends when the new session starts, so a
// full store gives its place to one reservation.
func (s *sessionStore) reserve(current string) (*place, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.full() {
		if !s.valid(current) || s.taken[current] {
			return nil, errNoPlace
		}

		s.taken[current] = true
	}

	s.reserved++

	return &place{store: s, replaces: current}, nil
}

func (s *sessionStore) valid(id string) bool {
	session, ok := s.sessions[id]

	return ok && s.now().Before(session.endsAt)
}

// free ends the reservation. The caller holds the lock of the store.
func (p *place) free() {
	p.store.reserved--
	delete(p.store.taken, p.replaces)
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

// start ends the session that the place replaces, and gives the ID of a new session in the place. The place is then
// used.
func (p *place) start(customerNumber string) string {
	id := newSessionID()

	p.once.Do(func() {
		s := p.store
		s.mu.Lock()
		defer s.mu.Unlock()

		p.free()
		delete(s.sessions, p.replaces)
		s.sessions[id] = session{customerNumber: customerNumber, endsAt: s.now().Add(sessionLifetime)}
	})

	return id
}

// release gives the place back and keeps the session that the place replaces, when no session started in it. A
// release after start does nothing.
func (p *place) release() {
	p.once.Do(func() {
		p.store.mu.Lock()
		defer p.store.mu.Unlock()

		p.free()
	})
}

func (s *sessionStore) customer(id string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.valid(id) {
		return "", false
	}

	return s.sessions[id].customerNumber, true
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
