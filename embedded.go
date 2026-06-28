package embedded

import "embed"

//go:embed all:web/dist
var Frontend embed.FS
