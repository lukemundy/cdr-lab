// Package buildinfo holds build facts that the build sets with -ldflags -X.
package buildinfo

var (
	commit      = "unknown"
	commitShort = "unknown"
	buildTime   = "unknown"
	id          = "local"
)

type Info struct {
	// Commit and CommitShort end in "-dirty" when the working tree had changes that were not committed.
	Commit      string
	CommitShort string
	Time        string
	ID          string
}

func Get() Info {
	return Info{Commit: commit, CommitShort: commitShort, Time: buildTime, ID: id}
}
