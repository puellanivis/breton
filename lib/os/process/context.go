package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
)

// signaler provides a basic api for channels to support optional signal handling, like for SIGHUP.
type signaler struct {
	sync.Once
	ch chan struct{}
}

// get returns a receive-only copy of the underlying channel for the signaler.
// If the channel does not exist, it will be allocated.
func (h *signaler) get() <-chan struct{} {
	h.Do(func() {
		h.ch = make(chan struct{})
	})

	return h.ch
}

// trySend does the job of trying to send the optional notification to the underlying channel.
// It will return true if it successfully sends a notification.
func (h *signaler) trySend() bool {
	select {
	case h.ch <- struct{}{}: // A send on a `nil` channel always blocks.
		return true
	default:
	}

	return false
}

var sigHandler struct {
	sync.Once
	ctx context.Context
	ch  chan os.Signal

	hup signaler
}

// HangupChannel provides a receiver channel that will notify the caller of when a SIGHUP signal has been received.
// This signal is often used by daemons to be notified of a request to reload configuration data.
//
// If the signal handler would block trzing to send this notification,
// then it will treat the signal the same as any other terminating signal.
func HangupChannel() <-chan struct{} {
	return sigHandler.hup.get()
}

// signalHandler starts a long-lived goroutine, which will run in the background until the process ends.
// It returns a `context.Context` that will be canceled in the event of a signal being received.
// It will then continue to listen for more signals,
// If it receives three signals, then we presume that we are not shutting down properly, and panic with all stacktraces.
func signalHandler(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)

	go func() {
		killChan := make(chan struct{}, 3)

		signal.Notify(sigHandler.ch, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
		for sig := range sigHandler.ch {
			fmt.Fprintln(os.Stderr, "received signal:", sig)

			switch sig {
			case syscall.SIGQUIT:
				debug.SetTraceback("all")
				panic("SIGQUIT")

			case syscall.SIGHUP:
				if sigHandler.hup.trySend() {
					continue
				}
			}

			cancel()

			select {
			case killChan <- struct{}{}: // killChan is not full, keep handling signals.
			default:
				// We have gotten three signals and we are still kicking,
				// panic and dump all stack traces.
				debug.SetTraceback("all")
				panic("not responding to signals")
			}
		}

		debug.SetTraceback("all")
		panic("signal handler channel unexpectedly closed")
	}()

	return ctx
}

// Context returns a process-level `context.Context`
// that is cancelled when the program receives a termination signal.
//
// A process should start a graceful shutdown process once this context is cancelled.
//
// This context should not be used as a parent to any requests,
// otherwise those requests will also be cancelled
// instead of being allowed to complete their work.
func Context() context.Context {
	sigHandler.Do(func() {
		sigHandler.ch = make(chan os.Signal, 1)
		sigHandler.ctx = signalHandler(context.Background())
	})

	return sigHandler.ctx
}

// Shutdown starts any graceful shutdown processes waiting for `process.Context()` to be cancelled.
//
// Shutdown works by injecting a `syscall.SIGTERM` directly to the signal handler,
// which will cancel the `process.Context()` the same as a real SIGTERM.
//
// Shutdown returns an error if it is unable to send the signal.
//
// Shutdown does not wait for anything to finish before returning.
func Shutdown() error {
	select {
	case sigHandler.ch <- syscall.SIGTERM:
		return nil
	default:
	}

	return errors.New("could not send signal")
}

// Quit ends the program as soon as possible, dumping a stacktrace of all goroutines.
//
// Quit works by injecting a `syscall.SIGQUIT` directly to the signal handler,
// which will cause a panic, and stacktrace of all goroutines the same as a real SIGQUIT.
//
// If Quit cannot inject the signal,
// it will setup an unrecoverable panic to occur.
//
// In all cases, Quit will not return.
func Quit() {
	select {
	case sigHandler.ch <- syscall.SIGQUIT:
	default:
		// We start up a separate goroutine for this to ensure that no `recover()` can block this panic.
		go func() {
			debug.SetTraceback("all")
			panic("process was force quit")
		}()
	}

	select {} // Block forever so we never return.
}
