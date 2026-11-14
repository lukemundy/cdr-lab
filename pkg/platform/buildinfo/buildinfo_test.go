package buildinfo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lukemundy/cdr-lab/pkg/platform/buildinfo"
)

func TestGetReturnsDefaultsWhenBuildSetsNothing(t *testing.T) {
	assert.Equal(t, buildinfo.Info{
		Commit:      "unknown",
		CommitShort: "unknown",
		Time:        "unknown",
		ID:          "local",
	}, buildinfo.Get())
}
