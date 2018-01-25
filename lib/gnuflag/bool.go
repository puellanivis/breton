package gnuflag

import (
	"strconv"
)

// -- bool Value
type boolValue bool

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	*b = boolValue(v)
	return err
}

func (b boolValue) Get() interface{} { return bool(b) }

func (b boolValue) String() string { return strconv.FormatBool(bool(b)) }

func (b boolValue) IsBoolFlag() bool { return true }

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type boolFlag interface {
	Value
	IsBoolFlag() bool
}

// Bool defines a bool flag with specified name, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func (f *FlagSet) Bool(name string, usage string, options ...Option) *bool {
	p := new(bool)
	f.Var((*boolValue)(p), name, usage, options...)
	return p
}

// Bool defines a bool flag with specified name, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func Bool(name string, usage string, options ...Option) *bool {
	return CommandLine.Bool(name, usage, options...)
}
