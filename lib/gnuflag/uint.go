package gnuflag

import (
	"strconv"
)

// -- uint Value
type uintValue uint

func (u *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	*u = uintValue(v)
	return err
}

func (u uintValue) Get() interface{} { return uint(u) }

func (u uintValue) String() string { return strconv.FormatUint(uint64(u), 10) }

// Uint defines a uint flag with specified name, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func (f *FlagSet) Uint(name string, usage string, options ...Option) *uint {
	p := new(uint)
	f.Var((*uintValue)(p), name, usage, options...)
	return p
}

// Uint defines a uint flag with specified name, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func Uint(name string, usage string, options ...Option) *uint {
	return CommandLine.Uint(name, usage, options...)
}

// -- uint64 Value
type uint64Value uint64

func (u *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	*u = uint64Value(v)
	return err
}

func (u uint64Value) Get() interface{} { return uint64(u) }

func (u uint64Value) String() string { return strconv.FormatUint(uint64(u), 10) }

// Uint64 defines a uint64 flag with specified name, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func (f *FlagSet) Uint64(name string, usage string, options ...Option) *uint64 {
	p := new(uint64)
	f.Var((*uint64Value)(p), name, usage, options...)
	return p
}

// Uint64 defines a uint64 flag with specified name, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64(name string, usage string, options ...Option) *uint64 {
	return CommandLine.Uint64(name, usage, options...)
}
