package templates

import "embed"

//go:embed layout.html index.html
var FS embed.FS
