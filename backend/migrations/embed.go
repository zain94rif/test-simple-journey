package migrations

import "embed"

// FS contains the versioned SQL migrations embedded into the binary.
//
//go:embed *.sql
var FS embed.FS
