package glog

import (
	"github.com/puellanivis/breton/lib/util"
)

func init() {
	util.AtExit(func() {
		Flush()
	})
}
