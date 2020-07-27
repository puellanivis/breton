package json

import (
	"encoding/json"
)

type config struct {
	*json.Encoder

	prefix, indent string

	compact bool
}

// An Option is a function that applies a specific option to an encoder config.
type Option func(*config)

// WithPrefix returns a function that directs Marshal to use the prefix string given.
func WithPrefix(prefix string) Option {
	return func(c *config) {
		c.prefix = prefix
	}
}

// WithIndent returns a function that directs Marshal to use the indenting string given.
func WithIndent(indent string) Option {
	return func(c *config) {
		c.indent = indent
	}
}

// EscapeHTML returns a function that directs Marshal to either enable or disable HTML escaping.
func EscapeHTML(on bool) Option {
	return func(c *config) {
		c.SetEscapeHTML(on)
	}
}

// Compact returns a function that directs Marshal to use compact format.
func Compact(value bool) Option {
	return func(c *config) {
		c.compact = value
	}
}
