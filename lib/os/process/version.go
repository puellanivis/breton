package process

var versionString string

// Version returns the version information populated during process.Init().
func Version() string {
	return versionString
}

func buildVersion(cmdname, semver, buildstamp string) {
	versionString = cmdname + " " + semver

	if buildstamp != "" {
		versionString = versionString + "-" + buildstamp
	}
}
