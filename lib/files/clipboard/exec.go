package clipboard

import (
	"bytes"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/puellanivis/breton/lib/files/wrapper"
)

type execClip struct {
	name  *url.URL
	paste []string
	copy  []string
}

func newExecClip(name string, target ...string) {
	if len(target) < 1 {
		target = append(target, name)
	}

	clipboards[name] = &execClip{
		name: &url.URL{
			Scheme: "clipboard",
			Opaque: url.PathEscape(name),
		},
		paste: append(pasteCmd, append(selParam, target...)...),
		copy:  append(copyCmd, append(selParam, target...)...),
	}
}

func (c *execClip) Read() ([]byte, error) {
	return exec.Command(c.paste[0], c.paste[1:]...).Output()
}

func (c *execClip) Write(b []byte) error {
	cmd := exec.Command(c.copy[0], c.copy[1:]...)

	cmd.Stdin = bytes.NewReader(b)

	return cmd.Run()
}

func (c *execClip) Stat() (os.FileInfo, error) {
	b, err := c.Read()
	if err != nil {
		return nil, err
	}

	return wrapper.NewInfo(c.name, len(b), time.Now()), nil
}
