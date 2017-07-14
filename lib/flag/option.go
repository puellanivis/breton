package flag

import (
	"fmt"
)

type Option func(*Flag) Option

func WithShort(shortFlag rune) Option {
	return func(f *Flag) Option {
		save := f.Short
		f.Short = shortFlag

		return WithShort(save)
	}
}

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
