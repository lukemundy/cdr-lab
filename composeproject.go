// Package cdrlab holds the compose project of the lab, so that the release mode of cdrlab up can write it into a
// directory with no checkout. The embed is at the root because the files are in four directories.
package cdrlab

import "embed"

//go:embed compose.yaml lab/certificates.json lab/postgres/*.sql services/register/seed/example.json
var ComposeProject embed.FS
