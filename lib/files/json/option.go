package json

import ()

type config struct {
	prefix, indent string
	compact        bool
}

// An Option is a function that apply a specific option, then returns an Option function
// that will revert the change applied.
type Option func(*config) Option

// WithIndent returns a function that directs Marshal to use the indenting characters given.
func WithIndent(prefix, indent string) Option {
	return func(c *config) Option {
		psave, isave := c.prefix, c.indent

		c.prefix, c.indent = prefix, indent

		return WithIndent(psave, isave)
	}
}

// Compact returns a function that directs Marshal to use compact format.
func Compact(value bool) Option {
	return func(c *config) Option {
		save := c.compact

		c.compact = value

		return Compact(save)
	}
}
