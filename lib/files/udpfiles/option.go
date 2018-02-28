package udpfiles

import (
	"github.com/puellanivis/breton/lib/files"
)

func WithIgnoreErrors(state bool) files.Option {
	type errorIgnorer interface{
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

func WithPacketSize(sz int) files.Option {
	type packetSizeSetter interface{
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
