package httpserver_test

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/pkg/platform/httpserver"
)

func listen(t *testing.T) (net.Listener, int) {
	t.Helper()

	l, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	return l, l.Addr().(*net.TCPAddr).Port
}

func freePort(t *testing.T) int {
	t.Helper()

	l, port := listen(t)
	require.NoError(t, l.Close())

	return port
}

func noContent(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }

func TestListenAndServeServesEachListenerUntilTheContextEnds(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	ports := []int{freePort(t), freePort(t)}
	stopped := make(chan error, 1)

	go func() {
		stopped <- httpserver.ListenAndServe(ctx, []httpserver.Listener{
			{Port: ports[0], Handler: http.HandlerFunc(noContent)},
			{Port: ports[1], Handler: http.HandlerFunc(noContent)},
		})
	}()

	for _, port := range ports {
		url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
			require.NoError(c, err)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(c, err)

			defer func() { _ = resp.Body.Close() }()

			assert.Equal(c, http.StatusNoContent, resp.StatusCode)
		}, 5*time.Second, 20*time.Millisecond)
	}

	cancel()

	assert.NoError(t, <-stopped)
}

func TestListenAndServeStopsWithAnErrorWhenAListenerFails(t *testing.T) {
	inUse, port := listen(t)
	t.Cleanup(func() { _ = inUse.Close() })

	err := httpserver.ListenAndServe(t.Context(), []httpserver.Listener{
		{Port: freePort(t), Handler: http.HandlerFunc(noContent)},
		{Port: port, Handler: http.HandlerFunc(noContent)},
	})

	assert.Error(t, err)
}

// Verifies: SRV-0014
func TestAnOpenRequestCompletesWhenTheServersStop(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	port := freePort(t)
	started := make(chan struct{})
	stopped := make(chan error, 1)

	go func() {
		stopped <- httpserver.ListenAndServe(ctx, []httpserver.Listener{{
			Port: port,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/slow" {
					close(started)
					time.Sleep(200 * time.Millisecond)
				}

				w.WriteHeader(http.StatusNoContent)
			}),
		}})
	}()

	baseURL := "http://127.0.0.1:" + strconv.Itoa(port)
	get := func(path string) (int, error) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, baseURL+path, nil)
		if err != nil {
			return 0, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}

		defer func() { _ = resp.Body.Close() }()

		return resp.StatusCode, nil
	}

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		code, err := get("/")
		require.NoError(c, err)
		assert.Equal(c, http.StatusNoContent, code)
	}, 5*time.Second, 20*time.Millisecond)

	status := make(chan int, 1)

	go func() {
		code, _ := get("/slow")
		status <- code
	}()

	<-started
	cancel()

	assert.Equal(t, http.StatusNoContent, <-status)
	assert.NoError(t, <-stopped)
}
