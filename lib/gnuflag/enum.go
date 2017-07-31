package gnuflag

import (
	"errors"
	"fmt"
	"strings"
)

// EnumValue describes a String flag that will only accept certain specific values.
type EnumValue struct {
	Value   int
	indices map[string]int
	valid   []string
}

// NewEnumValue returns an EnumValue that will only accept the given strings.
func NewEnumValue(valid []string) *EnumValue {
	e := &EnumValue{
		indices: make(map[string]int),
		valid:   valid,
	}

	for i, v := range valid {
		v = strings.ToUpper(v)

		e.indices[v] = i
	}

	return e
}

// ValueType permits an ability for the more standard flags library to support
// a flag displaying values/type allowed beyond the concrete types.
func (e *EnumValue) ValueType() string {
	return fmt.Sprintf("[ ", strings.Join(e.valid, ", "), " ]")
}

// Copy returns a newly allocated copy of the EnumValue.
func (e *EnumValue) Copy() interface{} {
	return &EnumValue{
		Value:   e.Value,
		indices: e.indices,
		valid:   e.valid,
	}
}

func newEnumValue(val string, p *EnumValue) *EnumValue {
	p.Set(val)
	return p
}

// String returns the canonical valid string of the value of the EnumValue.
func (e *EnumValue) String() string {
	if e.Value < 0 || e.Value >= len(e.valid) {
		return ""
	}
	return e.valid[e.Value]
}

// ErrBadEnum is the error returned when attempting to set an enum flag with a value not in the Enum
var ErrBadEnum = errors.New("bad enum value")

// Set attempts to set the given EnumValue to the given string.
func (e *EnumValue) Set(s string) error {
	if s == "" {
		e.Value = 0
		return nil
	}

	v, ok := e.indices[strings.ToUpper(s)]
	if !ok {
		return ErrBadEnum
	}

	e.Value = v
	return nil
}

// Get returns the value of the enum flag. Expect it to be of type int.
func (e *EnumValue) Get() interface{} {
	return e.Value
}

// EnumVar defines an enum flag with specified name, short flagname, usage, and default value. The argument p points to an EnumValue in which to store the value of the flag.
func (f *FlagSet) EnumVar(p *EnumValue, name string, value string, usage string, options ...Option) {
	f.Var(newEnumValue(value, p), name, usage, options...)
}

// EnumVar defines an enum flag with specified name, short flagname, usage, and default value. The argument p points to an EnumValue in which to store the value of the flag.
func EnumVar(p *EnumValue, name string, value string, usage string, options ...Option) {
	CommandLine.Var(newEnumValue(value, p), name, usage, options...)
}

// Enum defines an enum flag with specified name, shortname, usage, default value, and a list of additional valid values. The return value is the address of an EnumValue variable that stores the value of the flag.
func (f *FlagSet) Enum(name string, value string, usage string, valid []string, options ...Option) *EnumValue {
	e := NewEnumValue(valid)
	f.EnumVar(e, name, value, usage, options...)
	return e
}

// Enum defines an enum flag with specified name, shortname, usage, default value, and a list of additional valid values. The return value is the address of an EnumValue variable that stores the value of the flag.
func Enum(name string, value string, usage string, valid []string, options ...Option) *EnumValue {
	return CommandLine.Enum(name, value, usage, valid, options...)
}
