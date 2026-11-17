package httpserver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Verifies: SRV-0004, SRV-0005, SRV-0006, SRV-0007
func TestEachServerLimitsTheTimeOfARequestAndOfAnIdleConnection(t *testing.T) {
	srv := newServer(Listener{Port: 8080})

	assert.Equal(t, 5*time.Second, srv.ReadHeaderTimeout)
	assert.Equal(t, 30*time.Second, srv.ReadTimeout)
	assert.Equal(t, 60*time.Second, srv.WriteTimeout)
	assert.Equal(t, 120*time.Second, srv.IdleTimeout)
}
