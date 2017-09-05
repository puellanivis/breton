package gnuflag

import ()

// SetterFunc describes a function that takes a string from the command-line and performs some function that returns an error state.
type SetterFunc func(string) error

// FuncValue describes a flag which will call a func(string) error when specified as a flag.
type FuncValue struct {
	value  string
	isBool bool
	f      SetterFunc
}

// NewFunc returns a FuncValue that acts as a boolean flag.
func NewFunc(val string, fn func()) *FuncValue {
	return &FuncValue{
		value:  val,
		isBool: true,
		f:      func(s string) error { fn(); return nil },
	}
}

// NewFuncWithArg returns a FuncValue that acts as a normal flag.
func NewFuncWithArg(val string, fn SetterFunc) *FuncValue {
	return &FuncValue{
		value:  val,
		isBool: false,
		f:      fn,
	}
}

// String returns a String representation of this flag. (TODO: should be something other than empty string.
func (f *FuncValue) String() string {
	return f.value
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
func (f *FlagSet) FuncVar(fn *FuncValue, name, usage string, options ...Option) {
	f.Var(fn, name, usage, options...)
}

// FuncVar defines a function flag with specified name, shortname, and usage string. The argument p points to a FuncValue in which to store the function to call.
func FuncVar(fn *FuncValue, name, usage string, options ...Option) {
	CommandLine.Var(fn, name, usage, options...)
}

// Func defines a function flag with specified name, shortname, and usage string. It returns a pointer to a new FuncValue in which the function to call is stored.
func (f *FlagSet) Func(name, usage string, value func(), options ...Option) *FuncValue {
	fn := NewFunc("", value)
	f.FuncVar(fn, name, usage, options...)
	return fn
}

// FuncWithArg defines a function flag with specified name, shortname, and usage string. It returns a pointer to a new FuncValue in which the function to call is stored.
func (f *FlagSet) FuncWithArg(name, usage string, value SetterFunc, options ...Option) *FuncValue {
	fn := NewFuncWithArg("", value)
	f.FuncVar(fn, name, usage, options...)
	return fn
}

// Func defines a function flag with specified name, shortname, and usage string. It returns a pointer to a new FuncValue in which the function to call is stored.
func Func(name, usage string, value func(), options ...Option) *FuncValue {
	return CommandLine.Func(name, usage, value, options...)
}

// FuncWithArg defines a function flag with specified name, shortname, and usage string. It returns a pointer to a new FuncValue in which the function to call is stored.
func FuncWithArg(name, usage string, value SetterFunc, options ...Option) *FuncValue {
	return CommandLine.FuncWithArg(name, usage, value, options...)
}
