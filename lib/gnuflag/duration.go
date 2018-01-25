package gnuflag

import (
	"time"
)

// -- time.Duration Value
type durationValue time.Duration

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	*d = durationValue(v)
	return err
}

func (d durationValue) Get() interface{} { return time.Duration(d) }

func (d durationValue) String() string { return (time.Duration)(d).String() }

// Duration defines a time.Duration flag with specified name, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts values acceptable to time.ParseDuration.
func (f *FlagSet) Duration(name string, usage string, options ...Option) *time.Duration {
	p := new(time.Duration)
	f.Var((*durationValue)(p), name, usage, options...)
	return p
}

// Duration defines a time.Duration flag with specified name, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
// The flag accepts values acceptable to time.ParseDuration.
func Duration(name string, usage string, options ...Option) *time.Duration {
	return CommandLine.Duration(name, usage, options...)
}
