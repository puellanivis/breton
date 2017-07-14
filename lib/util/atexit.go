package util

import (
	"os"
	"runtime/pprof"
	"sync"

	"lib/flag"
	"lib/log"

	"context"
)

var (
	// Go libraries should not set flags themselves, this is an exception.
	profile = flag.String("profile", "", "write cpu profile to `filename`.prof and heap profile to filename.mprof")
)

// AtExitFunc is a function type that can be deferred for execution at the time of the program's exit. (Will NOT work if you call os.Exit(), you must call util.Exit())
type AtExitFunc func()

var (
	atExitMutex sync.Mutex
	atExitFuncs []AtExitFunc
)

// AtExit queues the given AtExitFuncs to be executed at program exit time.
func AtExit(f AtExitFunc) {
	atExitMutex.Lock()
	defer atExitMutex.Unlock()

	atExitFuncs = append([]AtExitFunc{f}, atExitFuncs...)
}

func runExitFuncs() {
	atExitMutex.Lock()
	defer atExitMutex.Unlock() // just in case

	for _, f := range atExitFuncs {
		f()
	}
}

// Exit runs the queued AtExitFuncs and then calls os.Exit with the given status.
func Exit(status int) {
	runExitFuncs()

	os.Exit(status)
}

// Init is initialization code that provides basic functionality for command-line programs. It parses flags, sets up AtExit, and queues flushing the log AtExit time. It takes as parameters version information, the first argument being the command's identifying string, and then a series of numbers which indicate the various version points. It returns a function that should be defer'ed in your main() function, which will take care of executing the queued AtExitFuncs even in a panic() situation.
func Init(cmdname string, versions ...int) AtExitFunc {
	initVersion(cmdname, versions...)

	flag.Parse()
	AtExit(func() {
		log.Flush()
	})

	if *profile != "" {
		f, err := os.Create(*profile + ".prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		AtExit(func() {
			pprof.StopCPUProfile()
			f.Close()

			f, err := os.Create(*profile + ".mprof")
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(f)
			f.Close()
		})
	}

	utilCtx = sigHandler(context.Background())

	return func() {
		runExitFuncs()

		if r := recover(); r != nil {
			panic(r)
		}
	}
}
