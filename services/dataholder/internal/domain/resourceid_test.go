package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lukemundy/cdr-lab/services/dataholder/internal/domain"
)

func TestAResourceIDIsAUUIDThatIsTheSameForTheSameInputOnly(t *testing.T) {
	id := domain.ResourceID("software-1", "00000001", "internal-1")

	parsed, err := uuid.Parse(id)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(8), parsed.Version())
	assert.Equal(t, uuid.RFC4122, parsed.Variant())
	assert.Equal(t, id, domain.ResourceID("software-1", "00000001", "internal-1"))

	for _, other := range []string{
		domain.ResourceID("software-2", "00000001", "internal-1"),
		domain.ResourceID("software-1", "00000002", "internal-1"),
		domain.ResourceID("software-1", "00000001", "internal-2"),
		// The length of each field is part of the input, so a different split of the same text is different.
		domain.ResourceID("software-1", "0000000", "1internal-1"),
	} {
		assert.NotEqual(t, id, other)
	}

	assert.NotContains(t, id, "software")
}
