package clipboard

import ()

var (
	pasteCmd = []string{"/usr/bin/pbpaste"}
	copyCmd  = []string{"/usr/bin/pbcopy"}
	selParam = []string{"-pboard"}
)

// special case, easy every time.
var Default Clipboard = &execClip{
	name:  ".",
	paste: pasteCmd,
	copy:  copyCmd,
}

func init() {
	clipboards["."] = Default

	newExecClip("general")
	newExecClip("ruler")
	newExecClip("find")
	newExecClip("font")
}
