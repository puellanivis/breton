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

	_ = LabelInvalid // use label in case it is commented above

	c := Counter("test", "testing counter", WithLabels(labels...))

	c.WithLabels(LabelFoo.WithValue("things")).Inc()
	c.WithLabels(LabelFoo.WithValue("stuff")).Add(42)
	c.WithLabels().Add(math.Pi)

	// all of these commented “tests“ are “undefined behavior”
	// as such actually testing for their behavior would define
	// their behavior, which is not the proper way to write tests.

	//c.WithLabels(LabelBar.WithValue("constant!")).Add(math.E)
	//c.WithLabels(LabelUndefined.WithValue("undefined!")).Add(math.Phi)
	//c.WithLabels(LabelUndefined).Add(math.Phi)

	_ = LabelUndefined // use label in case it is commented above

	t.Logf("counter: %#v", c)
	t.Logf("labels: %#v", c.labels.getList())
	t.Logf("labelset: %#v", c.labels.set)
	t.Logf("helpString: %q", c.helpString())
}
