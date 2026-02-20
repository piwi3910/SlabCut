// Package assets embeds static image resources into the binary.
package assets

import _ "embed"

//go:embed slabcut-icon.png
var IconPNG []byte

//go:embed slabcut-splash-1x.png
var SplashPNG []byte
