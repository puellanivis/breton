package gnuflag

import (
	"fmt"
)

// Option is a function that sets a specified function upon a flag.Flag.
// It returns an Option that will revert the option set.
type Option func(*Flag) Option

// WithShort returns an Option that will set the Short Flag of a flag to the
// specified rune.
func WithShort(shortFlag rune) Option {
	return func(f *Flag) Option {
		save := f.Short
		f.Short = shortFlag

		return WithShort(save)
	}
}

// WithDefault returns an Option that will set the default value of a flag
// to the value given. If when setting the value, the values cannot be assigned
// to the flag, then it will panic. (This should mostly be during initilization
// so while it cannot be checked at compile-time, it should result in an
// immediate panic when running the program.)
func WithDefault(value interface{}) Option {
	return func(f *Flag) Option {
		save := f.Value.Get()
		if err := f.Value.Set(fmt.Sprint(value)); err != nil {
			panic(err)
		}

		f.DefValue = f.Value.String()

		return WithDefault(save)
	}
}
