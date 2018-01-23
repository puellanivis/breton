package util

import (
	"fmt"
	"strconv"
	"strings"

	flag "github.com/puellanivis/breton/lib/gnuflag"
)

var (
	// Go libraries should not set flags themselves, this is an exception.
	_ = flag.Func("version", "display version information", func() {
		fmt.Println(versionString)
		Exit(0)
	})
)

var versionString string

// BUILD is a value that can be set by the go build command-line
// in order to provide additional context information for the build,
// such as build timestamp, branch, commit id, etc.
var BUILD = "adhoc"

// Version returns the version information populated during util.Init().
func Version() string {
	return versionString
}

func initVersion(cmdname string, versions ...uint) {
	var tmp []string

	if len(versions) < 1 {
		tmp = append(tmp, "0")
	}

	for _, ver := range versions {
		tmp = append(tmp, strconv.FormatUint(uint64(ver), 10))
	}

	versionString = fmt.Sprintf("%s v%s-%s", cmdname, strings.Join(tmp, "."), BUILD)
}
