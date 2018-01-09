package json

import ()

type config struct {
	prefix, indent string

	escapeHTML, compact bool
}

// An Option is a function that apply a specific option, then returns an Option function
// that will revert the change applied.
type Option func(*config) Option

// WithPrefix returns a function that directs Marshal to use the prefix string given.
func WithPrefix(prefix string) Option {
	return func(c *config) Option {
		save := c.prefix

		c.prefix = prefix

		return WithPrefix(save)
	}
}

// WithIndent returns a function that directs Marshal to use the indenting string given.
func WithIndent(indent string) Option {
	return func(c *config) Option {
		save := c.indent

		c.indent = indent

		return WithIndent(save)
	}
}

// EscapeHTML returns a function that directs Marshal to either enable or disable HTML escaping.
func EscapeHTML(value bool) Option {
	return func(c *config) Option {
		save := c.escapeHTML

		c.escapeHTML = value

		return EscapeHTML(save)
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
