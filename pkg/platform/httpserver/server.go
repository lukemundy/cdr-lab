package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

// Less than 10 seconds, the default wait before a container stop sends SIGKILL.
const shutdownTimeout = 5 * time.Second

type Listener struct {
	Port    int
	Handler http.Handler
	// TLSConfig is nil for plain HTTP.
	TLSConfig *tls.Config
}

// ListenAndServe serves all listeners until ctx ends or one listener fails, and then stops all of them.
func ListenAndServe(ctx context.Context, listeners []Listener) error {
	var servers []*http.Server

	serveErr := make(chan error, len(listeners))
	for _, l := range listeners {
		srv := newServer(l)

		servers = append(servers, srv)

		// net/http changes srv.TLSConfig when the server starts, so read it only before the start.
		useTLS := l.TLSConfig != nil
		go func() {
			if useTLS {
				serveErr <- srv.ListenAndServeTLS("", "")
				return
			}

			serveErr <- srv.ListenAndServe()
		}()

		slog.InfoContext(ctx, "server started", "server.port", l.Port, "tls", useTLS)
	}

	var err error
	select {
	case err = <-serveErr:
	case <-ctx.Done():
	}

	slog.InfoContext(ctx, "servers stopping")

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	for _, srv := range servers {
		err = errors.Join(err, srv.Shutdown(shutdownCtx))
	}

	return err
}

func newServer(l Listener) *http.Server {
	return &http.Server{
		Addr:              ":" + strconv.Itoa(l.Port),
		Handler:           l.Handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		TLSConfig:         l.TLSConfig,
	}
}
