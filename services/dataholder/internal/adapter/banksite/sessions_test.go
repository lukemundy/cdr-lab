package banksite

import (
	"encoding/base64"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(limit int) (*sessionStore, *time.Time) {
	clock := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)

	return newSessionStore(limit, func() time.Time { return clock }), &clock
}

func start(t *testing.T, store *sessionStore, customerNumber string) string {
	t.Helper()

	place, err := store.reserve()
	require.NoError(t, err)

	return place.start(customerNumber)
}

// Verifies: DH-2035
func TestASessionIDHas256RandomBits(t *testing.T) {
	store, _ := newTestStore(10)

	first, second := start(t, store, "00000001"), start(t, store, "00000001")

	raw, err := base64.RawURLEncoding.DecodeString(first)
	require.NoError(t, err)
	assert.Len(t, raw, 32)
	assert.NotEqual(t, first, second)
}

// Verifies: DH-2033, DH-2044
func TestASessionGivesItsCustomerUntilItEnds(t *testing.T) {
	store, _ := newTestStore(10)
	id := start(t, store, "00000001")

	customer, ok := store.customer(id)
	require.True(t, ok)
	assert.Equal(t, "00000001", customer)

	store.end(id)

	_, ok = store.customer(id)
	assert.False(t, ok)
}

// Verifies: DH-2038, DH-2039
func TestASessionEnds24HoursAfterItsStartWithNoInactivityTimeout(t *testing.T) {
	store, clock := newTestStore(10)
	id := start(t, store, "00000001")

	*clock = clock.Add(23*time.Hour + 59*time.Minute)
	_, ok := store.customer(id)
	assert.True(t, ok)

	*clock = clock.Add(time.Minute)
	_, ok = store.customer(id)
	assert.False(t, ok)
}

// Verifies: DH-2041
func TestAFullStoreRefusesANewSessionAndKeepsTheSessionsThatDidNotEnd(t *testing.T) {
	store, _ := newTestStore(2)
	first := start(t, store, "00000001")
	start(t, store, "00000002")

	_, err := store.reserve()
	require.ErrorIs(t, err, errNoPlace)

	_, ok := store.customer(first)
	assert.True(t, ok)
}

// Verifies: DH-2041
func TestAFullStoreRemovesTheSessionsThatEndedBeforeItRefuses(t *testing.T) {
	store, clock := newTestStore(2)
	start(t, store, "00000001")

	*clock = clock.Add(12 * time.Hour)
	second := start(t, store, "00000002")
	*clock = clock.Add(12 * time.Hour)

	start(t, store, "00000003")

	_, ok := store.customer(second)
	assert.True(t, ok)
}

func TestAReservationTakesAPlaceUntilItStartsOrIsReleased(t *testing.T) {
	store, _ := newTestStore(1)

	place, err := store.reserve()
	require.NoError(t, err)
	_, err = store.reserve()
	require.ErrorIs(t, err, errNoPlace)

	place.release()
	place.release()

	place, err = store.reserve()
	require.NoError(t, err)
	place.start("00000001")
	place.release()

	_, err = store.reserve()
	assert.ErrorIs(t, err, errNoPlace)
}

// Verifies: DH-2041
func TestParallelReservationsDoNotGoAboveTheLimit(t *testing.T) {
	store, _ := newTestStore(50)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		granted int
	)

	for range 200 {
		wg.Go(func() {
			if _, err := store.reserve(); err == nil {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		})
	}

	wg.Wait()
	assert.Equal(t, 50, granted)
}
