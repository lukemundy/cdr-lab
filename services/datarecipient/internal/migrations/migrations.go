// Package migrations contains the goose SQL migrations of the data recipient database.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
