package aboutfiles

import (
	"time"

	"github.com/puellanivis/breton/lib/os/process"
)

type stringFunc func() string

func (f stringFunc) ReadAll() ([]byte, error) {
	return append([]byte(f()), '\n'), nil
}

var (
	blank   stringFunc = func() string { return "" }
	version stringFunc = func() string { return process.Version() }
	now     stringFunc = func() string { return time.Now().Truncate(0).String() }
)
