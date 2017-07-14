package flag

import (
)

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func (f *FlagSet) StringVar(p *string, name string, usage string, value string, options ...Option) {
	f.Var(newStringValue(value, p), name, usage, options...)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, usage string, value string, options ...Option) {
	CommandLine.Var(newStringValue(value, p), name, usage, options...)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func (f *FlagSet) String(name string, usage string, value string, options ...Option) *string {
	p := new(string)
	f.StringVar(p, name, usage, value, options...)
	return p
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, usage string, value string, options ...Option) *string {
	return CommandLine.String(name, usage, value, options...)
}
