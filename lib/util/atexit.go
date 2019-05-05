package util

import (
	"context"

	"github.com/puellanivis/breton/lib/os/process"
)

// AtExitFunc is a function type that can be deferred for execution at the time of the program's exit. (Will NOT work if you call os.Exit(), you must call util.Exit())
type AtExitFunc func()

// AtExit queues the given AtExitFuncs to be executed at program exit time.
func AtExit(f AtExitFunc) {
	process.AtExit(f)
}

// Exit runs the queued AtExitFuncs and then calls os.Exit with the given status.
func Exit(status int) {
	process.Exit(status)
}

// Init is initialization code that provides basic functionality for command-line programs.
// It parses flags, sets up AtExit, and will start profiling if the appropriate flag is set.
// It takes as parameters version information,
// the first argument being the command's identifying string,
// and then a series of numbers which indicate the various version points.
// It returns the context.Context from util.Context(),
// and a function that should be defer'ed in your main() function,
// which will take care of executing the queued AtExitFuncs even in a panic() situation.
func Init(cmdname string, versions ...uint) (context.Context, func()) {
	return process.Init(cmdname, buildSemver(versions), BUILD)
}
