package labdir

import (
	"testing"

	"github.com/lukemundy/cdr-lab/pkg/platform/database/databasetest"
)

// Verifies: PKI-0001
func TestIssue(t *testing.T) {
	t.Parallel()

	// Verifies: PKI-0001, PKI-0002
	t.Run("revoke", func(t *testing.T) {})
}
