package util

import (
	"context"
	"os"
	"runtime/pprof"
	"sync"

	flag "github.com/puellanivis/breton/lib/gnuflag"
)

var (
	// Go libraries should not set flags themselves, this is an exception.
	profile = flag.String("profile", "write cpu profile to `filename`.prof and heap profile to filename.mprof")
)

// AtExitFunc is a function type that can be deferred for execution at the time of the program's exit. (Will NOT work if you call os.Exit(), you must call util.Exit())
type AtExitFunc func()

var atExit struct {
	sync.Mutex
	f []AtExitFunc
}

// AtExit queues the given AtExitFuncs to be executed at program exit time.
func AtExit(f AtExitFunc) {
	atExit.Lock()
	defer atExit.Unlock()

	atExit.f = append([]AtExitFunc{f}, atExit.f...)
}

func runExitFuncs() {
	atExit.Lock()
	defer atExit.Unlock() // just in case

	for _, f := range atExit.f {
		f()
	}
}

// Exit runs the queued AtExitFuncs and then calls os.Exit with the given status.
func Exit(status int) {
	runExitFuncs()

	os.Exit(status)
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
	initVersion(cmdname, versions...)

	flag.Parse()

	if *profile != "" {
		cpuf, err := os.Create(*profile + ".prof")
		if err != nil {
			panic(err)
		}

		memf, err := os.Create(*profile + ".mprof")
		if err != nil {
			panic(err)
		}

		_ = pprof.StartCPUProfile(cpuf)

		AtExit(func() {
			pprof.StopCPUProfile()
			cpuf.Close()

			_ = pprof.WriteHeapProfile(memf)
			memf.Close()
		})
	}

	return utilCtx, func() {
		runExitFuncs()

		if r := recover(); r != nil {
			panic(r)
		}
	}
}
