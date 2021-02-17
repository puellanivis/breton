package process

var versionString string

// Version returns the full version information populated during process.Init().
//
// This function is intended to implement a `--version` flag, and thus returns `"cmdname version[-buildstamp]"`.
// It does _not_ return just the `"version[-buildstamp]"` information, for that use `AppVersion()`.
func Version() string {
	return versionString
}

var (
	appName    string
	appVersion string
)

// AppName returns the command name populated during process.Init().
func AppName() string {
	return appName
}

// AppVersion returns only the version and buildstamp populated during process.Init().
func AppVersion() string {
	return appVersion
}

func buildVersion(cmdname, version, buildstamp string) {
	appName, appVersion = cmdname, version
	if buildstamp != "" {
		appVersion += "-" + buildstamp
	}

	versionString = appName + " " + appVersion
}
