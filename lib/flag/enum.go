package flag

import (
	"errors"
	"fmt"
	"strings"
)

type EnumValue struct {
	Value   int
	indices map[string]int
	valid   []string
}

func NewEnumValue(valid ...string) *EnumValue {
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

func (e *EnumValue) ValueType() string {
	return fmt.Sprintf("[ ", strings.Join(e.valid, ", "), " ]")
}

func (e *EnumValue) Copy() interface{} {
	return &EnumValue{
		Value:   e.Value,
		indices: e.indices,
		valid:   e.valid,
	}
}

func newEnum(val string, p *EnumValue) *EnumValue {
	p.Set(val)
	return p
}

func (e *EnumValue) String() string {
	if e.Value < 0 || e.Value >= len(e.valid) {
		return ""
	}
	return e.valid[e.Value]
}

// ErrBadEnum is the error returned when attempting to set an enum flag with a value not in the Enum
var ErrBadEnum = errors.New("bad enum value")

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

func (e *EnumValue) Get() interface{} {
	return e.Value
}

func (f *FlagSet) EnumVar(e *EnumValue, name string, short rune, usage string, value string) {
	f.Var(newEnum(value, e), name, short, usage)
}

func EnumVar(e *EnumValue, name string, short rune, usage string, value string) {
	CommandLine.Var(newEnum(value, e), name, short, usage)
}

func (f *FlagSet) Enum(name string, short rune, usage string, value string, valid ...string) *EnumValue {
	e := NewEnumValue(valid...)
	f.EnumVar(e, name, short, usage, value)
	return e
}

func Enum(name string, short rune, usage string, value string, valid ...string) *EnumValue {
	return CommandLine.Enum(name, short, usage, value, valid...)
}
