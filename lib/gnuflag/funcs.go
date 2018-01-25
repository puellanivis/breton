package gnuflag

import (
	"fmt"
)

// setterFunc describes a function that takes a string from the command-line and performs some function that returns an error state.
type setterFunc func(string) error

// FuncValue describes a flag which will call a func(string) error when specified as a flag.
type funcValue struct {
	name, value string
	isBool      bool

	f setterFunc
}

// newBoolFunc returns a FuncValue that acts as a boolean flag.
func newBoolFunc(name string, fn func()) *funcValue {
	return &funcValue{
		name:   name,
		isBool: true,
		f:      func(s string) error { fn(); return nil },
	}
}

// newFunc returns a FuncValue that acts as a normal flag.
func newFunc(name string, fn func(string) error) *funcValue {
	return &funcValue{
		name:   name,
		isBool: false,
		f:      fn,
	}
}

// String returns a String representation of this flag.
func (f *funcValue) String() string {
	return fmt.Sprintf("%s(%q)", f.name, f.value)
}

// Set calls the function of the FuncValue and returns its error.
func (f *funcValue) Set(s string) error {
	f.value = s
	return f.f(s)
}

// Get returns the underlying `func(string) error` function.
func (f *funcValue) Get() interface{} {
	return f.f
}

// IsBoolFlag implements the test for if a flag should act as a boolean flag.
func (f *funcValue) IsBoolFlag() bool {
	return f.isBool
}

// BoolFunc defines a function flag with specified name, and usage string.
// It returns a pointer to the niladic function.
func (f *FlagSet) BoolFunc(name, usage string, value func(), options ...Option) func() {
	fn := newBoolFunc(name, value)
	f.Var(fn, name, usage, options...)
	return value
}

// Func defines a function flag with specified name, and usage string.
// It returns a pointer to the SetterFunc.
func (f *FlagSet) Func(name, usage string, value func(string) error, options ...Option) func(string) error {
	fn := newFunc(name, value)
	f.Var(fn, name, usage, options...)
	return value
}

// BoolFunc defines a function flag with specified name, and usage string.
// It returns a pointer to the niladic function.
func BoolFunc(name, usage string, value func(), options ...Option) func() {
	return CommandLine.BoolFunc(name, usage, value, options...)
}

// Func defines a function flag with specified name, shortname, and usage string.
// It returns a pointer to the SetterFunc.
func Func(name, usage string, value func(string) error, options ...Option) func(string) error {
	return CommandLine.Func(name, usage, value, options...)
}
