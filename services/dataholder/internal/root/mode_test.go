package root

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/adapter/statusstore"
	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func TestNoSubcommandIsServe(t *testing.T) {
	got, found := modeOf(nil)

	require.True(t, found)
	assert.Equal(t, mode{api: true, workers: true}, got)
}

func TestEachModeSubcommandIsAMode(t *testing.T) {
	tests := map[string]mode{
		"serve":  {api: true, workers: true},
		"api":    {api: true},
		"worker": {workers: true},
	}
	for subcommand, want := range tests {
		t.Run(subcommand, func(t *testing.T) {
			got, found := modeOf([]string{subcommand})

			require.True(t, found)
			assert.Equal(t, want, got)
		})
	}

	for _, args := range [][]string{{"migrate"}, {"serve", "api"}, {"workers"}} {
		_, found := modeOf(args)
		assert.False(t, found, args)
	}
}

// Verifies: DH-2622, DH-2625
func TestEachModeHasOnlyItsServersAndItsWorkers(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	allPorts := []int{internalPort, apiPort, securePort}

	tests := map[string]struct {
		mode  mode
		ports []int
		works bool
	}{
		"serve":  {mode: mode{api: true, workers: true}, ports: allPorts, works: true},
		"api":    {mode: mode{api: true}, ports: allPorts},
		"worker": {mode: mode{workers: true}, ports: []int{internalPort}, works: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setOnly(t, tc.mode)

			cfg, err := loadConfig(tc.mode)
			require.NoError(t, err)

			got, err := newProcess(t.Context(), pool, cfg, tc.mode)
			require.NoError(t, err)

			ports := make([]int, 0, len(got.listeners))
			for _, listener := range got.listeners {
				ports = append(ports, listener.Port)
			}

			assert.Equal(t, tc.ports, ports)
			assert.Equal(t, tc.works, got.jobs != nil)
		})
	}
}

func TestAProcessWithTheAPIsIsReadyOnlyAfterTheFirstStatuses(t *testing.T) {
	setOnly(t, serveMode)

	cfg, err := loadConfig(serveMode)
	require.NoError(t, err)

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	ready := func(m mode) int {
		proc, err := newProcess(t.Context(), pool, cfg, m)
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		proc.listeners[0].Handler.ServeHTTP(
			recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil),
		)

		return recorder.Code
	}

	assert.Equal(t, http.StatusServiceUnavailable, ready(mode{api: true, workers: true}))
	assert.Equal(t, http.StatusServiceUnavailable, ready(mode{api: true}))
	assert.Equal(t, http.StatusOK, ready(mode{workers: true}), "a worker gets the statuses")

	require.NoError(t, statusstore.New(pool, nil).Save(t.Context(), map[string]domain.SoftwareProductStatus{}, `"v1"`))

	assert.Equal(t, http.StatusOK, ready(mode{api: true, workers: true}))
	assert.Equal(t, http.StatusOK, ready(mode{api: true}))
}
