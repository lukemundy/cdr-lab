package eventstore

import _ "embed"

// Schema creates the events table. Each service that has a Store copies it into a migration of its own.
//
//go:embed events.sql
var Schema string
