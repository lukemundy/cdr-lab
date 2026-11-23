package buildinfo

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReturnsDefaultsWhenBuildSetsNothing(t *testing.T) {
	assert.Equal(t, Info{
		Version:     "dev",
		Commit:      "unknown",
		CommitShort: "unknown",
		Time:        "unknown",
		ID:          "local",
	}, Get())
}

func flags() values {
	return values{version: "dev", commit: "unknown", commitShort: "unknown", time: "unknown", id: "local"}
}

func goBuild(version string, settings ...debug.BuildSetting) *debug.BuildInfo {
	return &debug.BuildInfo{Main: debug.Module{Version: version}, Settings: settings}
}

const revision = "0123456789abcdef0123456789abcdef01234567"

// Verifies: CLI-0115
func TestAVersionThatTheBuildSetsIsTheVersion(t *testing.T) {
	v := flags()
	v.version = "rehearsal-42"

	assert.Equal(t, "rehearsal-42", resolve(v, goBuild("v0.3.1")).Version)
}

// Verifies: CLI-0116, CLI-0117
func TestWithNoVersionFromTheBuildTheVersionIsTheModuleVersionOnlyForAReleaseVersion(t *testing.T) {
	for module, want := range map[string]string{
		"v0.3.1":                               "v0.3.1",
		"v12.0.10":                             "v12.0.10",
		"v0.3.2-0.20261201030000-0123456789ab": "dev",
		"v0.3.1+dirty":                         "dev",
		"v0.3.1-rc.1":                          "dev",
		"v01.2.3":                              "dev",
		"(devel)":                              "dev",
		"":                                     "dev",
	} {
		t.Run(module, func(t *testing.T) {
			assert.Equal(t, want, resolve(flags(), goBuild(module)).Version)
		})
	}
}

func TestWithNoGoBuildInformationTheVersionIsDev(t *testing.T) {
	assert.Equal(t, "dev", resolve(flags(), nil).Version)
}

func TestWithNoCommitFromTheBuildTheCommitIsTheVCSRevision(t *testing.T) {
	info := resolve(flags(), goBuild("", debug.BuildSetting{Key: "vcs.revision", Value: revision}))

	assert.Equal(t, revision, info.Commit)
	assert.Equal(t, "0123456", info.CommitShort)
}

func TestAVCSRevisionWithChangesEndsInDirty(t *testing.T) {
	info := resolve(flags(), goBuild("",
		debug.BuildSetting{Key: "vcs.revision", Value: revision},
		debug.BuildSetting{Key: "vcs.modified", Value: "true"},
	))

	assert.Equal(t, revision+"-dirty", info.Commit)
	assert.Equal(t, "0123456-dirty", info.CommitShort)
}

func TestACommitFromTheBuildIsKept(t *testing.T) {
	v := flags()
	v.commit, v.commitShort = "fedcba", "fed"

	info := resolve(v, goBuild("", debug.BuildSetting{Key: "vcs.revision", Value: revision}))

	assert.Equal(t, "fedcba", info.Commit)
	assert.Equal(t, "fed", info.CommitShort)
}

// Verifies: CLI-0118
func TestWithNoCommitFromAnySourceTheCommitIsUnknown(t *testing.T) {
	info := resolve(flags(), goBuild("v0.3.1"))

	assert.Equal(t, "unknown", info.Commit)
	assert.Equal(t, "unknown", info.CommitShort)
}

// Verifies: CLI-0143
func TestIsReleaseVersionAcceptsOnlyAReleaseVersion(t *testing.T) {
	for version, want := range map[string]bool{
		"v0.3.1":       true,
		"v10.0.12":     true,
		"0.3.1":        false,
		"v0.3":         false,
		"v0.3.1-rc.1":  false,
		"v01.2.3":      false,
		"dev":          false,
		"rehearsal-42": false,
	} {
		assert.Equal(t, want, IsReleaseVersion(version), version)
	}
}
