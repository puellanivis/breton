package util

import (
	"context"

	"github.com/puellanivis/breton/lib/os/process"
)

// HangupChannel provides a receiver channel that will notify the caller of when a SIGHUP signal has been received.
// This signal is often used by daemons to be notified of a request to reload configuration data.
//
// Note, that if too many SIGHUPs arrive at once, and the signal handler would block trying to send this notification,
// then it will treat the signal the same as any other terminating signals.
func HangupChannel() <-chan struct{} {
	return process.HangupChannel()
}

// Context returns a context.Context that will cancel if the program receives any signal that a program may want to cleanup after.
func Context() context.Context {
	return process.Context()
}

// Quit causes a closure of the signal channel, which causes the signal handler
// to cancel the util.Context() context, leading to a (hopefully) graceful-ish
// shutdown of the program.
func Quit() {
	process.Quit()
}
