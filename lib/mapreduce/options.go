package mapreduce

import (
)

type config struct{
	threadCount int
	mapperCount int
	stripeSize int

	ordered bool
}

type Option func(c *config) Option

func WithThreadCount(num int) Option {
	return func(c *config) Option {
		save := c.threadCount

		c.threadCount = num

		return WithThreadCount(save)
	}
}

func WithMapperCount(num int) Option {
	return func(c *config) Option {
		save := c.mapperCount

		c.mapperCount = num

		return WithMapperCount(save)
	}
}

func WithStripeSize(size int) Option {
	return func(c *config) Option {
		save := c.stripeSize

		c.stripeSize = size

		return WithStripeSize(save)
	}
}

func WithOrdering(ordered bool) Option {
	return func(c *config) Option {
		save := c.ordered

		c.ordered = ordered

		return WithOrdering(save)
	}
}
