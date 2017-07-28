package clipboard

import (
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"
)

var (
	osdir    = os.Getenv("SystemRoot") + "\\System32\\"
	pasteCmd = []string{osdir + "WindowsPowerShell\\v1.0\\powershell.exe", "Get-Clipboard"}
	copyCmd  = []string{osdir + "clip.exe"}
	selParam = []string{}
)

const (
	cfUnicodetext = 13
)

var (
	user32           = syscall.MustLoadDLL("user32")
	openClipboard    = user32.MustFindProc("OpenClipboard")
	closeClipboard   = user32.MustFindProc("CloseClipboard")
	getClipboardData = user32.MustFindProc("GetClipboardData")

	kernel32     = syscall.NewLazyDLL("kernel32")
	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")
)

type winClip struct{}

var defaultClipboard clipboard = winClip{}

func init() {
	clipboards["."] = defaultClipboard
}

func (c winClip) Read() ([]byte, error) {
	r, _, err := openClipboard.Call(0)
	if r == 0 {
		return nil, err
	}
	defer closeClipboard.Call()

	h, _, err := getClipboardData.Call(cfUnicodetext)
	if r == 0 {
		return nil, err
	}

	l, _, err := globalLock.Call(h)
	if l == 0 {
		return nil, err
	}
	defer globalUnlock.Call(h)

	text := syscall.UTF16ToString((*[1 << 20]uint16)(unsafe.Pointer(l))[:])

	return []byte(text), nil
}

func (c winClip) Write(b []byte) error {
	cmd := exec.Command(copyCmd[0])

	wr, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := wr.Write(b); err != nil {
		_ = wr.Close() // explicitly throwing away error here
		return err
	}

	if err := wr.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

func (c winClip) Name() string {
	return "."
}

func (c winClip) Size() int64 {
	b, _ := c.Read()
	return int64(len(b))
}

func (c winClip) Mode() os.FileMode {
	return os.ModePerm
}

func (c winClip) ModTime() time.Time {
	return time.Now()
}

func (c winClip) IsDir() bool {
	return false
}

func (c winClip) Sys() interface{} {
	return nil
}
