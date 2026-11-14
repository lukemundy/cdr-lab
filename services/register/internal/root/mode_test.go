package root

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
	"github.com/lukemundy/cdr-lab/services/register/internal/app"
)

var (
	serveMode  = mode{api: true, workers: true}
	apiMode    = mode{api: true}
	workerMode = mode{workers: true}
)

func TestNoSubcommandIsServe(t *testing.T) {
	got, found := modeOf(nil)

	require.True(t, found)
	assert.Equal(t, serveMode, got)
}

func TestEachModeSubcommandIsAMode(t *testing.T) {
	for subcommand, want := range map[string]mode{"serve": serveMode, "api": apiMode, "worker": workerMode} {
		got, found := modeOf([]string{subcommand})

		require.True(t, found, subcommand)
		assert.Equal(t, want, got, subcommand)
	}

	for _, args := range [][]string{{"migrate"}, {"serve", "api"}, {"workers"}} {
		_, found := modeOf(args)
		assert.False(t, found, args)
	}
}

// onlyCASettings removes each setting that a process with no APIs does not need.
func onlyCASettings(t *testing.T) {
	t.Helper()

	setLabSettings(t)

	for _, name := range []string{
		"PUBLIC_BASE_URL", "SECURE_BASE_URL", "OPERATOR_TOKEN", "OPERATOR_ALLOWED_HOSTS", "PUBLIC_TLS_CERT_FILE",
		"PUBLIC_TLS_KEY_FILE", "SECURE_TLS_CERT_FILE", "SECURE_TLS_KEY_FILE",
	} {
		t.Setenv(name, "")
	}
}

func TestAWorkerNeedsOnlyTheSettingsOfTheCA(t *testing.T) {
	onlyCASettings(t)

	_, err := loadConfig(workerMode)
	require.NoError(t, err)

	_, err = loadConfig(apiMode)
	require.Error(t, err)

	t.Setenv("CDR_INTERMEDIATE_KEY_FILE", "")

	_, err = loadConfig(workerMode)
	assert.ErrorContains(t, err, "CDR_INTERMEDIATE_KEY_FILE")
}

func TestEachModeHasOnlyItsServersAndItsWorkers(t *testing.T) {
	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	allPorts := []int{probePort, apiPort, operatorPort, securePort, crlPort}

	tests := map[string]struct {
		mode  mode
		ports []int
		works bool
	}{
		"serve":  {mode: serveMode, ports: allPorts, works: true},
		"api":    {mode: apiMode, ports: allPorts},
		"worker": {mode: workerMode, ports: []int{probePort}, works: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			setLabSettings(t)

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

func TestAProcessWithTheAPIsIsReadyOnlyWithACurrentCRL(t *testing.T) {
	setLabSettings(t)

	cfg, err := loadConfig(serveMode)
	require.NoError(t, err)

	pool := databasetest.New(t)
	require.NoError(t, migrate(t.Context(), pool))

	ready := func(m mode) int {
		t.Helper()

		proc, err := newProcess(t.Context(), pool, cfg, m)
		require.NoError(t, err)

		rec := httptest.NewRecorder()
		proc.listeners[0].Handler.ServeHTTP(
			rec,
			httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil),
		)

		return rec.Code
	}

	assert.Equal(t, http.StatusOK, ready(workerMode), "a worker does not serve the CRL")
	assert.Equal(t, 0, crlRows(t, pool), "a worker signs a CRL only in its job")

	assert.Equal(t, http.StatusOK, ready(apiMode), "the start signs a CRL")

	proc, err := newProcess(t.Context(), pool, cfg, serveMode)
	require.NoError(t, err)

	_, err = pool.Exec(t.Context(), "UPDATE cdr_crl SET next_update = now() - interval '1 second'")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	proc.listeners[0].Handler.ServeHTTP(
		rec,
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/readyz", nil),
	)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestTheRenewalIntervalKeepsTheCRLRenewalWithin24Hours(t *testing.T) {
	assert.LessOrEqual(t, app.CRLRenewalAge+crlRenewalInterval, 24*time.Hour)
}
