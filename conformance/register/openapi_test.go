package register

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThePinnedSpecificationLoads(t *testing.T) {
	_, err := LoadSpec()

	require.NoError(t, err)
}
