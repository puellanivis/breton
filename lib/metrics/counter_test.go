package metrics

import (
	"math"
	"testing"
)

const (
	LabelFoo       = Label("foo")
	LabelBar       = Label("bar")
	LabelInvalid   = Label("invalid!")
	LabelUndefined = Label("undefined")
)

func TestCounter(t *testing.T) {
	labels := []Labeler{
		LabelFoo.WithValue("default"),
		LabelBar.Const("stuff"),
		//LabelInvalid,
	}

	c := Counter("test", "testing counter", WithLabels(labels...))

	c.WithLabels(LabelFoo.WithValue("things")).Inc()
	c.WithLabels(LabelFoo.WithValue("stuff")).Add(42)
	c.WithLabels().Add(math.Pi)

	//c.WithLabels(LabelBar.WithValue("constant!")).Add(math.E)
	//c.WithLabels(LabelUndefined.WithValue("undefined!")).Add(math.Phi)
	//c.WithLabels(LabelUndefined).Add(math.Phi)

	t.Logf("counter: %#v", c)
	t.Logf("labels: %#v", c.labels.getList())
	t.Logf("labelset: %#v", c.labels.set)
}
