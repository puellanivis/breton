package util

import (
	"fmt"
	"strings"

	"lib/flag"
)

var (
	// Go libraries should not set flags themselves, this is an exception.
	_ = flag.Func("version", 0, "display version information", func() {
		fmt.Println(versionString)
		Exit(0)
	})
)

var versionString string

// BUILD is a value that can be set by the go build command-line in order to provide an additional context information for the build.
var BUILD = "adhoc"

// Version returns the version information populated during util.Init().
func Version() string {
	return versionString
}

func initVersion(cmdname string, versions ...int) {
	var tmp []string

	if len(versions) < 1 {
		tmp = []string{"0"}
	}

	for _, ver := range versions {
		tmp = append(tmp, fmt.Sprint(ver))
	}

	versionString = fmt.Sprintf("%s v%s-%s", cmdname, strings.Join(tmp, "."), BUILD)
}
