package json

import (
)

type config struct {
	prefix, indent string
	compact bool
}

type Option func(*config) Option

func Indent(prefix, indent string) Option {
	return func(c *config) Option {
		psave, isave := c.prefix, c.indent

		c.prefix, c.indent = prefix, indent

		return Indent(psave, isave)
	}
}

func Compact(value bool) Option {
	return func(c *config) Option {
		save := c.compact

		c.compact = value

		return Compact(save)
	}
}
