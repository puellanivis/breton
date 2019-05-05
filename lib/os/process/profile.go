package process

import (
	"os"
	"runtime/pprof"
)

func setupProfiling() {
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
