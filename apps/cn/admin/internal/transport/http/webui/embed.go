package webui

import "embed"

// FS embeds the React production frontend written to dist/ by apps/cn/admin/react.
//
//go:embed dist
var FS embed.FS
