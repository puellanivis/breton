package socketfiles

import (
	"github.com/puellanivis/breton/lib/files"
)

// WithIgnoreErrors means that a write to a given files.File will silently drop errors.
//
// Requires the files.File to implement `interface{ IgnoreErrors(bool) bool }`, or else no action is taken.
//
// Really only useful for writing to Broadcast or UDP addresses before they are opened by a listener.
func WithIgnoreErrors(state bool) files.Option {
	type errorIgnorer interface {
		IgnoreErrors(bool) bool
	}

	return func(f files.File) (files.Option, error) {
		var save bool

		if w, ok := f.(errorIgnorer); ok {
			save = w.IgnoreErrors(state)
		}

		return WithIgnoreErrors(save), nil
	}
}

// WithPacketSize chunks each Write to a specified size.
//
// Requires the files.File to implement `interface{ SetPacketSize(int) int }`, or else no action is taken.
func WithPacketSize(sz int) files.Option {
	type packetSizeSetter interface {
		SetPacketSize(int) int
	}

	return func(f files.File) (files.Option, error) {
		var save int

		if w, ok := f.(packetSizeSetter); ok {
			save = w.SetPacketSize(sz)
		}

		return WithPacketSize(save), nil
	}
}
