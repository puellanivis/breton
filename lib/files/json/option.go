package json

import (
)

type config struct {
	prefix, indent string
	compact bool
}

type option func(*config) option

func Indent(prefix, indent string) option {
	return func(c *config) option {
		psave, isave := c.prefix, c.indent

		c.prefix, c.indent = prefix, indent

		return Indent(psave, isave)
	}
}

func Compact(value bool) option {
	return func(c *config) option {
		save := c.compact

		c.compact = value

		return Compact(save)
	}
}
