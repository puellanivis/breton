package util

import (
	"strconv"
	"strings"

	"github.com/puellanivis/breton/lib/os/process"
)

// BUILD is a value that can be set by the go build command-line
// in order to provide additional context information for the build,
// such as build timestamp, branch, commit id, etc.
var BUILD = "adhoc"

// Version returns the version information populated during util.Init().
func Version() string {
	return process.Version()
}

func buildSemver(versions []uint) string {
	var tmp []string

	if len(versions) < 1 {
		tmp = append(tmp, "0")
	}

	for _, ver := range versions {
		tmp = append(tmp, strconv.FormatUint(uint64(ver), 10))
	}

	return "v" + strings.Join(tmp, ".")
}
