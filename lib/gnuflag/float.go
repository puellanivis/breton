package gnuflag

import (
	"strconv"
)

// -- float64 Value
type float64Value float64

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float64Value(v)
	return err
}

func (f float64Value) Get() interface{} { return float64(f) }

func (f float64Value) String() string { return strconv.FormatFloat(float64(f), 'g', -1, 64) }

// Float defines a float64 flag with specified name, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (f *FlagSet) Float(name string, usage string, options ...Option) *float64 {
	p := new(float64)
	if err := f.Var((*float64Value)(p), name, usage, options...); err != nil {
		panic(err)
	}
	return p
}

// Float defines a float64 flag with specified name, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float(name string, usage string, options ...Option) *float64 {
	return CommandLine.Float(name, usage, options...)
}
