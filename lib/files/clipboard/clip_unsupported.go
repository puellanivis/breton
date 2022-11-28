//go:build android || nacl || plan9 || zos
// +build android nacl plan9 zos

package clipboard

var (
	pasteCmd []string
	copyCmd  []string
	selParam []string
)
