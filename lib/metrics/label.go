package metrics

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/puellanivis/breton/lib/metrics/internal/kv"
)

// cannot use `const LabelVar = Label("asdf")` if Label is a function that
// tests at compile-time. So, we will check at Registration time.
var validLabel = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type Labeler interface {
	// Label returns the name=value pairing from the implementer.
	Label() (name, value string)
}

// Label describes a label name.
type Label string

// Label implements Labeler.
func (l Label) Label() (name, value string) {
	return string(l), ""
}

// WithLabel takes the given label name, and attaches a value to it.
func (l Label) WithValue(value string) Labeler {
	return LabelValue{
		key: string(l),
		val: value,
	}
}

// Const returns a Labeler that defines a label with a constant value.
// The concrete type is unexported in order to further enforce constantness.
func (l Label) Const(value string) Labeler {
	return constLabel{
		key: string(l),
		val: value,
	}
}

// A LabelValue is a pair of Label name, and a value.
type LabelValue struct {
	key, val string
}

// Label implements Labeler.
func (l LabelValue) Label() (name, value string) {
	return l.key, l.val
}

// WithLabel takes the given label name, and attaches a value to it.
func (l LabelValue) WithValue(value string) Labeler {
	return LabelValue{
		key: l.key,
		val: value,
	}
}

type constLabel LabelValue

// Label implements Labeler.
func (l constLabel) Label() (name, value string) {
	return l.key, l.val
}

// WithLabel takes the given label name, and attaches a value to it.
func (l constLabel) WithValue(value string) Labeler {
	panic(fmt.Sprintf("attempt to assign to constant label %q", l.key))
}

// labelSet describes a set of labels, i.e. which keys are valid, and which are constant.
type labelSet struct {
	keys   []string
	canSet map[string]bool
}

func newLabelSet(labels []Labeler) *labelSet {
	s := &labelSet{
		canSet: make(map[string]bool),
	}

	for _, label := range labels {
		k, _ := label.Label()

		if !validLabel.MatchString(k) {
			panic(fmt.Sprintf("label name %q is invalid", k))
		}

		if _, ok := s.canSet[k]; ok {
			panic(fmt.Sprintf("label %q redefined", k))
		}

		// Let’s assume it can be set
		s.canSet[k] = true

		if _, ok := label.(constLabel); ok {
			// Well, ok then, it is a constant.
			s.canSet[k] = false
		}
	}

	s.keys = nil // fastest way to empty the list
	for key := range s.canSet {
		s.keys = append(s.keys, key)
	}
	sort.Strings(s.keys)

	return s
}

type labelScope struct {
	set *labelSet // keep track of the labelSet, for canSet testing
	p   *labelScope   // keep track of the parent of this scope

	kv kv.KeyVal // the key:val set defined at this scope.
}

func defineLabels(labels ...Labeler) *labelScope {
	l := &labelScope{
		set: newLabelSet(labels),
	}

	for _, label := range labels {
		if k, v := label.Label(); v != "" {
			l.kv.Append(k, v)
		}
	}
	sort.Sort(l.kv)

	return l
}

// WithLabels returns a child labelScope object that additionally has the given Labelers labels set.
func (l *labelScope) With(labels ...Labeler) *labelScope {
	if l == nil {
		panic("metric does not have any labels")
	}

	n := &labelScope{
		set: l.set,
		p:   l,
	}

	for _, label := range labels {
		k, v := label.Label()

		canSet, ok := n.set.canSet[k]
		if !ok {
			panic(fmt.Sprintf("attempt to assign to undefined label %q", k))
		}

		if !canSet {
			panic(fmt.Sprintf("attempt to assign to constant label %q", k))
		}

		n.kv.Append(k, v)
	}
	sort.Sort(n.kv)

	return n
}

func (l *labelScope) get(name string) Labeler {
	if i, ok := l.kv.Index(name); ok {
		return LabelValue{
			key: name,
			val: l.kv.Vals[i],
		}
	}

	if l.p != nil {
		return l.p.get(name)
	}

	return Label(name)
}

// Get returns a slice of Labelers that defines the labels and their values.
func (l *labelScope) getList() []Labeler {
	var list []Labeler

	for _, k := range l.set.keys {
		list = append(list, l.get(k))
	}

	return list
}

// getMap is a conversion function necessary to wrap the prometheus client.
func (l *labelScope) getMap() map[string]string {
	m := make(map[string]string)

	for _, k := range l.set.keys {
		_, v := l.get(k).Label()

		m[k] = v
	}

	return m
}
