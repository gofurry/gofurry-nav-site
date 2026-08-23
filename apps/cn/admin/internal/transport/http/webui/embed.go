package webui

import "embed"

// FS embeds the built frontend (dist/).
// The build pipeline should copy the frontend build output into this folder.
//
//go:embed dist
var FS embed.FS
