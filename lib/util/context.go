package util

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/puellanivis/breton/lib/glog"
)

var (
	utilCtx context.Context
	sigchan = make(chan os.Signal)
)

// Context returns a context.Context that will cancel if the program receives any signal that a program may want to cleanup after.
func Context() context.Context {
	return utilCtx
}

// IsDone is a helper function that without blocking returns true/false
// if the context is done. (Makes it easier to just, “if done { return }"
func IsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
	}

	return false
}

// Quit causes a closure of the signal channel, which causes the signal handler
// to cancel the util.Context() context, leading to a (hopefully) graceful-ish
// shutdown of the program.
func Quit() {
	select {
	case <-sigchan:
		return
	default:
	}

	// by closing this channel, we cause all reads from sigchan to now
	// succeed, which means, we essentially make a kill signal, but
	// without actually causing a signal. (Important for Windows)
	close(sigchan)
}

func sigHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	killchan := make(chan struct{}, 3)
	sigchan = make(chan os.Signal)

	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for {
			select {
			case sig := <-sigchan:
				glog.Error("received signal:", sig)

				if !IsDone(ctx) {
					cancel()
				}

				select {
				case killchan <- struct{}{}:
					go func() {
						<-time.After(1 * time.Second)

						select {
						case <-killchan:
						default:
						}
					}()

				default:
					debug.SetTraceback("all")
					panic("not responding to signals!")
				}
			}
		}
	}()

	return ctx
}
