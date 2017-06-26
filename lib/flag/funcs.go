// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package flag

import ()

type FuncValue struct {
	name   string
	isBool bool
	f      func(string)
}

func NewFunc(name string, fn func()) *FuncValue {
	return &FuncValue{
		name:   name,
		isBool: true,
		f:      func(_ string) { fn() },
	}
}

func NewFuncWithArg(name string, fn func(string)) *FuncValue {
	return &FuncValue{
		name:   name,
		isBool: false,
		f:      fn,
	}
}

func (f *FuncValue) String() string {
	return ""
}

func (f *FuncValue) Set(s string) error {
	f.f(s)
	return nil
}

func (f *FuncValue) Get() interface{} {
	return f.f
}

func (f *FuncValue) IsBoolFlag() bool {
	return f.isBool
}

func (f *FlagSet) FuncVar(fn *FuncValue, name string, short rune, usage string) {
	f.Var(fn, name, short, usage)
}

func FuncVar(fn *FuncValue, name string, short rune, usage string) {
	CommandLine.Var(fn, name, short, usage)
}

func (f *FlagSet) Func(name string, short rune, usage string, value func()) *FuncValue {
	fn := NewFunc(name, value)
	f.FuncVar(fn, name, short, usage)
	return fn
}

func (f *FlagSet) FuncWithArg(name string, short rune, usage string, value func(string)) *FuncValue {
	fn := NewFuncWithArg(name, value)
	f.FuncVar(fn, name, short, usage)
	return fn
}

func Func(name string, short rune, usage string, value func()) *FuncValue {
	return CommandLine.Func(name, short, usage, value)
}

func FuncWithArg(name string, short rune, usage string, value func(string)) *FuncValue {
	return CommandLine.FuncWithArg(name, short, usage, value)
}
