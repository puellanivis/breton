package process

import (
	"context"
	"fmt"

	flag "github.com/puellanivis/breton/lib/gnuflag"
)

// Go libraries should not set flags themselves, these are an exception.
var (
	profile = flag.String("profile", "write cpu profile to `filename`.prof and heap profile to filename.mprof")
	_       = flag.BoolFunc("version", "display version information", func() {
		fmt.Println(Version())
		Exit(0)
	})
)

// Init is initialization code that provides basic functionality for command-line programs.
// It parses flags, sets up AtExit, and will start profiling if the appropriate flag is set.
// It takes as parameters version information,
// the first argument being the command's identifying string,
// and then a series of numbers which indicate the various version points.
// It returns the context.Context from util.Context(),
// and a function that should be defer'ed in your main() function,
// which will take care of executing the queued AtExitFuncs even in a panic() situation.
func Init(cmdname, semver, buildstamp string) (context.Context, func()) {
	buildVersion(cmdname, semver, buildstamp)

	flag.Parse()

	if *profile != "" {
		setupProfiling()
	}

	return Context(), func() {
		runExitFuncs()

		if r := recover(); r != nil {
			panic(r)
		}
	}
}
