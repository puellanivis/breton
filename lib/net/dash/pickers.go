package dash

import (
	"github.com/puellanivis/breton/lib/net/dash/mpd"
)

// A PickRepFunc takes two DASH MPD Representations, and returns true
// if the cur Representation should replace the best Representation.
// Note: best can be nil
type PickRepFunc func(best, cur *mpd.Representation) bool

// PickFirst returns true when best == nil (the first test), and false thereafter.
func PickFirst(best, cur *mpd.Representation) bool {
	return best == nil
}

// PickHighestBandwidth returns true if cur.Bandwidth is greater than the best.
func PickHighestBandwidth(best, cur *mpd.Representation) bool {
	if best == nil {
		return true
	}

	return best.Bandwidth < cur.Bandwidth
}

// PickLowestBandwidth returns true if cur.Bandwidth is less than the best.
func PickLowestBandwidth(best, cur *mpd.Representation) bool {
	if best == nil {
		return true
	}

	return cur.Bandwidth < best.Bandwidth
}
