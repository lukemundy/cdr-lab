// Package migrations contains the goose SQL migrations of the data holder database.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
