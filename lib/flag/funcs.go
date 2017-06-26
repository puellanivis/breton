// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package flag

import ()

// SetterFunc describes a function that takes a string from the command-line and performs some function that returns an error state.
type SetterFunc	func(string) error

// FuncValue describes a flag which will call a func(string) error when specified as a flag.
type FuncValue struct {
	name   string
	isBool bool
	f      SetterFunc
}

// NewFunc returns a FuncValue that acts as a boolean flag.
func NewFunc(name string, fn func()) *FuncValue {
	return &FuncValue{
		name:   name,
		isBool: true,
		f:      func(_ string) error { fn(); return nil },
	}
}

// NewFuncWithArg returns a FuncValue that acts as a normal flag.
func NewFuncWithArg(name string, fn SetterFunc) *FuncValue {
	return &FuncValue{
		name:   name,
		isBool: false,
		f:      fn,
	}
}

// String returns a String representation of this flag. (TODO: should be something other than empty string.
func (f *FuncValue) String() string {
	return ""
}

// Set calls the function of the FuncValue and returns its error.
func (f *FuncValue) Set(s string) error {
	return f.f(s)
}

// Get returns the underlying function.
func (f *FuncValue) Get() interface{} {
	return f.f
}

// IsBoolFlag implements the test for if a flag should act as a boolean flag.
func (f *FuncValue) IsBoolFlag() bool {
	return f.isBool
}

// FuncVar defines a function flag with specified name, shortname, and usage string. The argument p points to a FuncValue in which to store the function to call.
func (f *FlagSet) FuncVar(fn *FuncValue, name string, short rune, usage string) {
	f.Var(fn, name, short, usage)
}

// FuncVar defines a function flag with specified name, shortname, and usage string. The argument p points to a FuncValue in which to store the function to call.
func FuncVar(fn *FuncValue, name string, short rune, usage string) {
	CommandLine.Var(fn, name, short, usage)
}

// Func defines a function flag with specified name, shortname, and usage string. It returns a pointer to a new FuncValue in which the function to call is stored.
func (f *FlagSet) Func(name string, short rune, usage string, value func()) *FuncValue {
	fn := NewFunc(name, value)
	f.FuncVar(fn, name, short, usage)
	return fn
}

// FuncWithArg defines a function flag with specified name, shortname, and usage string. It returns a pointer to a new FuncValue in which the function to call is stored.
func (f *FlagSet) FuncWithArg(name string, short rune, usage string, value SetterFunc) *FuncValue {
	fn := NewFuncWithArg(name, value)
	f.FuncVar(fn, name, short, usage)
	return fn
}

// Func defines a function flag with specified name, shortname, and usage string. It returns a pointer to a new FuncValue in which the function to call is stored.
func Func(name string, short rune, usage string, value func()) *FuncValue {
	return CommandLine.Func(name, short, usage, value)
}

// FuncWithArg defines a function flag with specified name, shortname, and usage string. It returns a pointer to a new FuncValue in which the function to call is stored.
func FuncWithArg(name string, short rune, usage string, value SetterFunc) *FuncValue {
	return CommandLine.FuncWithArg(name, short, usage, value)
}
