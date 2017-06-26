package edge

import (
	"testing"
)

func TestEdge(t *testing.T) {
	var g Edge

	if !g.Up() {
		t.Error("edge didn't trigger")
	}

	if g.Up() {
		t.Error("edge triggered twice")
	}

	if !g.Down() {
		t.Error("edge didn't trigger down")
	}

	if g.Down() {
		t.Error("edge triggered down twice")
	}

	if !g.Up() {
		t.Error("edge didn't retrigger after triggering down")
	}
}
