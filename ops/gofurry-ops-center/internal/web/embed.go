package web

import "embed"

// FS embeds the built ops dashboard.
//
//go:embed dist
var FS embed.FS
