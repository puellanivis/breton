package util

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

var (
	utilCtx context.Context
	sigchan = make(chan os.Signal)

	hupMutex sync.Mutex
	hupchan  chan struct{}
)

// HangupChannel provides a receiver channel that will notify the caller of when a SIGHUP signal has been received.
// This signal is often used by daemons to be notified of a request to reload configuration data.
//
// Note, that if too many SIGHUPs arrive at once, and the signal handler would block trying to send this notification,
// then it will treat the signal the same as any other terminating signals.
func HangupChannel() <-chan struct{} {
	hupMutex.Lock()
	defer hupMutex.Unlock()

	if hupchan == nil {
		hupchan = make(chan struct{})
	}

	return hupchan
}

// sendHup does the job of locking the hupMutex and trying to send the HUP notification.
// It will return true if it successfully sends the notification.
func sendHup() bool {
	hupMutex.Lock()
	defer hupMutex.Unlock()

	if hupchan == nil {
		return false
	}

	select {
	case hupchan <- struct{}{}:
		return true
	default:
		fmt.Fprintln(os.Stderr, "too many hangungs: terminating")
	}

	return false
}

func init() {
	var cancel context.CancelFunc
	utilCtx, cancel = context.WithCancel(context.Background())

	signal.Notify(sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		killchan := make(chan struct{}, 3)

		for sig := range sigchan {
			fmt.Fprintln(os.Stderr, "received signal:", sig)

			switch sig {
			case syscall.SIGHUP:
				if sendHup() {
					continue
				}
			}

			cancel()

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
				panic("not responding to signals")
			}
		}

		debug.SetTraceback("all")
		panic("process was force quit")
	}()
}

// Context returns a context.Context that will cancel if the program receives any signal that a program may want to cleanup after.
func Context() context.Context {
	return utilCtx
}

// Quit causes a closure of the signal channel, which causes the signal handler
// to cancel the util.Context() context, leading to a (hopefully) graceful-ish
// shutdown of the program.
func Quit() {
	select {
	case _, closed := <-sigchan:
		if closed {
			return
		}
	default:
	}

	// by closing this channel, the for range signchan above ends,
	// which means, we essentially make a kill signal, but
	// without actually causing a signal.
	// (Important for Windows, which doesn’t have signaling.)
	close(sigchan)
}
