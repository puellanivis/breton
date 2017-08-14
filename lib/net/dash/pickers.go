package dash

import (
	"github.com/puellanivis/breton/lib/net/dash/mpd"
)

// A Picker is a function that will take a series of DASH MPD Representations,
// and returns either the given Representation, or an alternative,
// if the given Representation is not appropriate by rules of the Picker.
// After all Representations have been given to this function, the return value
// will be the best-fitting Representation.
type Picker func(cur *mpd.Representation) *mpd.Representation

// PickFirst returns a Picker that will always return the first Representation passed into it.
func PickFirst() Picker {
	var best *mpd.Representation

	return func(cur *mpd.Representation) *mpd.Representation {
		if best == nil {
			best = cur
		}

		return best
	}
}

// PickHighestBandwidth returns a picker that selects the Representation with the highest Bandwidth.
// Between two equal bandwidths, the first is picked.
func PickHighestBandwidth() Picker {
	var best *mpd.Representation

	return func(cur *mpd.Representation) *mpd.Representation {
		if best == nil || best.Bandwidth < cur.Bandwidth {
			best = cur
		}

		return best
	}
}

// PickLowestBandwidth returns a picker that selects the Representation with the lowest Bandwidth.
// Between two equal bandwidths, the first is picked.
func PickLowestBandwidth() Picker {
	var best *mpd.Representation

	return func(cur *mpd.Representation) *mpd.Representation {
		if best == nil || best.Bandwidth > cur.Bandwidth {
			best = cur
		}

		return best
	}
}
