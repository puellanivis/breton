package glog

import (
	"github.com/puellanivis/breton/lib/os/process"
)

func init() {
	process.AtExit(func() {
		Flush()
	})
}
