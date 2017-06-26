package clipboard

import (
	"os"
	"os/exec"
	"time"
)

type execClip struct {
	name string
	paste []string
	copy []string
}

func newExecClip(name string, target ...string) {
	if len(target) < 1 {
		target = append(target, name)
	}

	clipboards[name] = &execClip{
		name: name,
		paste: append(pasteCmd, append(selParam, target...)...),
		copy: append(copyCmd, append(selParam, target...)...),
	}
}

func (c *execClip) Read() ([]byte, error) {
	return exec.Command(c.paste[0], c.paste[1:]...).Output()
}

func (c *execClip) Write(b []byte) error {
	cmd := exec.Command(c.copy[0], c.copy[1:]...)

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

func (c *execClip) Name() string {
	return c.name
}

func (c *execClip) Size() int64 {
	b, _ := c.Read()
	return int64(len(b))
}

func (c *execClip) Mode() os.FileMode {
	return os.ModePerm
}

func (c *execClip) ModTime() time.Time {
	return time.Now()
}

func (c *execClip) IsDir() bool {
	return c.name == "."
}

func (c *execClip) Sys() interface{} {
	return nil
}
