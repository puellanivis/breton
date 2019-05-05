package gnuflag

import (
	"strconv"
)

// -- int Value
type intValue int

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	*i = intValue(v)
	return err
}

func (i intValue) Get() interface{} { return int(i) }

func (i intValue) String() string { return strconv.Itoa(int(i)) }

// Int defines an int flag with specified name, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func (f *FlagSet) Int(name string, usage string, options ...Option) *int {
	p := new(int)
	if err := f.Var((*intValue)(p), name, usage, options...); err != nil {
		panic(err)
	}
	return p
}

// Int defines an int flag with specified name, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, usage string, options ...Option) *int {
	return CommandLine.Int(name, usage, options...)
}

// -- int64 Value
type int64Value int64

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i int64Value) Get() interface{} { return int64(i) }

func (i int64Value) String() string { return strconv.FormatInt(int64(i), 10) }

// Int64 defines an int64 flag with specified name, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func (f *FlagSet) Int64(name string, usage string, options ...Option) *int64 {
	p := new(int64)
	if err := f.Var((*int64Value)(p), name, usage, options...); err != nil {
		panic(err)
	}
	return p
}

// Int64 defines an int64 flag with specified name, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64(name string, usage string, options ...Option) *int64 {
	return CommandLine.Int64(name, usage, options...)
}
