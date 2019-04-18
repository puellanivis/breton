package process

import (
	"fmt"
	"strings"
)

var versionString string

// Version returns the version information populated during util.Init().
func Version() string {
	return versionString
}

func buildVersion(cmdname, semver, buildstamp string) {
	semver = strings.TrimPrefix(semver, "v")

	versionString = fmt.Sprintf("%s v%s-%s", cmdname, semver, buildstamp)
}
