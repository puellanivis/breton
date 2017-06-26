package clipboard

import (
	"os/exec"
)

var (
	pasteCmd []string
	copyCmd []string
	selParam []string
)

var (
	xclipPaste = []string{"-out" }
	xclipCopy = []string{"-in" }
)

var (
	xselPaste = []string{"--output"}
	xselCopy = []string{"--input"}
)

var Default Clipboard

func init() {
	if cmd, err := exec.LookPath("xclip"); err == nil {
		pasteCmd = append([]string{cmd}, xclipPaste...)
		copyCmd = append([]string{cmd}, xclipCopy...)
		selParam = []string{ "-selection" }

		newExecClip(".", "clipboard")
		Default = clipboards["."]

		newExecClip("clipboard")
		newExecClip("primary")
		newExecClip("secondary")
		return
	}

	cmd, err := exec.LookPath("xsel")
	if err != nil {
		return
	}

	pasteCmd = append([]string{cmd}, xselPaste...)
	copyCmd = append([]string{cmd}, xselCopy...)
	selParam = []string{}

	newExecClip(".", "--clipboard")
	Default = clipboards["."]

	newExecClip("clipboard", "--clipboard")
	newExecClip("primary", "--primary")
	newExecClip("secondary", "--secondary")
}
