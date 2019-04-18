package process

import (
	"os"
	"sync"
)

var atExit struct {
	sync.Mutex
	f []func()
}

// AtExit registers a function to be called when `process.Exit` from this package is called.
//
// Subsequent calls to AtExit do not overwrite previous calls,
// and registered functions are executed in stack order,
// in order to mimic the behavior of `defer`.
//
// Since AtExit does not hook into the standard-library `os.Exit`,
// you must avoid using any function that calls `os.Exit` (most often `Fatal`-type logging methods).
func AtExit(f func()) {
	atExit.Lock()
	defer atExit.Unlock()

	atExit.f = append(atExit.f, f)
}

func runExitFuncs() {
	atExit.Lock()
	// intentionally do not unlock.

	for i := len(atExit.f) - 1; i >= 0; i-- {
		atExit.f[i]()
	}
}

// Exit causes the current program to exit with the given status code.
//
// Exit runs the sequence of functions established by `process.AtExit`,
// and then calls `os.Exit` with the given status.
func Exit(status int) {
	runExitFuncs()

	os.Exit(status)
}
