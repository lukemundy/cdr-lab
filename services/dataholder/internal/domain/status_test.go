package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func TestOnlyTheStatusesOfASoftwareProductOnTheRegisterAreKnown(t *testing.T) {
	for _, status := range []string{"ACTIVE", "INACTIVE", "REMOVED"} {
		got, known := domain.ParseSoftwareProductStatus(status)

		assert.True(t, known, status)
		assert.Equal(t, domain.SoftwareProductStatus(status), got)
	}

	for _, status := range []string{"", "active", "SUSPENDED", "ACTIVE "} {
		_, known := domain.ParseSoftwareProductStatus(status)

		assert.False(t, known, status)
	}
}
