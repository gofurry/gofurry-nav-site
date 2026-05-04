package web

import "embed"

// FS embeds the built admin console.
//
//go:embed dist
var FS embed.FS
