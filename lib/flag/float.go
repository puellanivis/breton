package flag

import (
	"strconv"
)

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() interface{} { return float64(*f) }

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

// FloatVar defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func (f *FlagSet) FloatVar(p *float64, name string, usage string, value float64, options ...Option) {
	f.Var(newFloat64Value(value, p), name, usage, options...)
}

// FloatVar defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func FloatVar(p *float64, name string, usage string, value float64, options ...Option) {
	CommandLine.Var(newFloat64Value(value, p), name, usage, options...)
}

// Float defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func (f *FlagSet) Float(name string, usage string, value float64, options ...Option) *float64 {
	p := new(float64)
	f.FloatVar(p, name, usage, value, options...)
	return p
}

// Float defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float(name string, usage string, value float64, options ...Option) *float64 {
	return CommandLine.Float(name, usage, value, options...)
}
