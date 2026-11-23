// Package buildinfo holds build facts that the build sets with -ldflags -X. When the build sets no commit or no
// version, it uses the Go build information of the binary.
package buildinfo

import (
	"regexp"
	"runtime/debug"
)

var (
	// version is only for the cdrlab CLI. The services give the commit as their version.
	version     = "dev"
	commit      = "unknown"
	commitShort = "unknown"
	buildTime   = "unknown"
	id          = "local"
)

type Info struct {
	// Version is a version that the build sets, or a release version of the Go module, or "dev".
	Version string
	// Commit and CommitShort end in "-dirty" when the working tree had changes that were not committed.
	Commit      string
	CommitShort string
	Time        string
	ID          string
}

func Get() Info {
	bi, _ := debug.ReadBuildInfo()

	return resolve(values{version: version, commit: commit, commitShort: commitShort, time: buildTime, id: id}, bi)
}

type values struct {
	version, commit, commitShort, time, id string
}

// releaseVersion is the format of a release version (tools/release/DESIGN.md#rls-0004).
var releaseVersion = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

func IsReleaseVersion(v string) bool {
	return releaseVersion.MatchString(v)
}

// shortLength is the length of an abbreviated commit of git describe.
const shortLength = 7

func resolve(v values, bi *debug.BuildInfo) Info {
	info := Info{Version: v.version, Commit: v.commit, CommitShort: v.commitShort, Time: v.time, ID: v.id}
	if bi == nil {
		return info
	}

	if info.Version == "dev" && IsReleaseVersion(bi.Main.Version) {
		info.Version = bi.Main.Version
	}

	if info.Commit == "unknown" {
		setVCSCommit(&info, bi.Settings)
	}

	return info
}

func setVCSCommit(info *Info, settings []debug.BuildSetting) {
	var revision, suffix string

	for _, s := range settings {
		switch {
		case s.Key == "vcs.revision":
			revision = s.Value
		case s.Key == "vcs.modified" && s.Value == "true":
			suffix = "-dirty"
		}
	}

	if len(revision) < shortLength {
		return
	}

	info.Commit = revision + suffix
	info.CommitShort = revision[:shortLength] + suffix
}
