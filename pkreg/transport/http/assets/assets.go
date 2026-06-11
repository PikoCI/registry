package assets

import "embed"

//go:embed all:css all:js all:images all:fonts
var Assets embed.FS
