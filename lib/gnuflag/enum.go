package gnuflag

import (
	"errors"
	"strconv"
	"strings"
)

// EnumValue defines an string-like flag type that maps strings to uint values.
// It is exported, so that it may be defined in a gnuflags.Struct parameter.
type EnumValue int

// enumValue describes a String flag that will only accept certain specific values.
type enumValue struct {
	val     *int
	indices map[string]int
	valid   []string
}

// newEnumValue returns an EnumValue that will only accept the given strings.
func newEnumValue(valid ...string) *enumValue {
	e := &enumValue{
		val: new(int),
	}

	e.setValid(valid)

	return e
}

func (e *enumValue) setValid(valid []string) {
	e.valid = valid

	e.indices = make(map[string]int)
	for i, v := range valid {
		if v == "" {
			continue
		}

		v = strings.ToUpper(v)

		e.indices[v] = i
	}
}

// ValueType permits an ability for the more standard flags library to support
// a flag displaying values/type allowed beyond the concrete types.
func (e *enumValue) ValueType() string {
	var filtered []string

	for _, v := range e.valid {
		if v == "" {
			continue
		}

		filtered = append(filtered, v)
	}

	return strings.Join(filtered, ", ")
}

// String returns the canonical valid string of the value of the EnumValue.
func (e *enumValue) String() string {
	var val int
	if e.val != nil {
		val = *e.val
	}

	if val < 0 || val >= len(e.valid) {
		return strconv.Itoa(val)
	}

	return e.valid[val]
}

// ErrBadEnum is the error returned when attempting to set an enum flag with a value not in the Enum
var ErrBadEnum = errors.New("bad enum value")

// Set attempts to set the given EnumValue to the given string.
func (e *enumValue) Set(s string) error {
	if e.val == nil {
		return errors.New("uninitialized enum usage")
	}

	if s == "" {
		*e.val = 0
		return nil
	}

	v, ok := e.indices[strings.ToUpper(s)]
	if !ok {
		return ErrBadEnum
	}

	*e.val = v
	return nil
}

// Get returns the value of the enum flag. Expect it to be of type int.
func (e *enumValue) Get() interface{} {
	if e.val == nil {
		return 0
	}

	return *e.val
}

// Enum defines an enum flag with specified name, usage, and list of valid values.
// The return value is the address of an EnumValue variable that stores the value of the flag.
func (f *FlagSet) Enum(name string, usage string, valid []string, options ...Option) *EnumValue {
	e := newEnumValue(valid...)
	f.Var(e, name, usage, options...)
	return (*EnumValue)(e.val)
}

// Enum defines an enum flag with specified name, usage, and list of valid values.
// The return value is the address of an EnumValue variable that stores the value of the flag.
func Enum(name string, usage string, valid []string, options ...Option) *EnumValue {
	return CommandLine.Enum(name, usage, valid, options...)
}
