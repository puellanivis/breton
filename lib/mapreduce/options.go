package mapreduce

type config struct {
	threadCount int
	mapperCount int
	stripeSize  int

	ordered bool
}

// Option defines a function that applies a setting or value to a MapReduce configuration.
// It returns a function that will undo the setting or value that was applied.
type Option func(c *config) Option

// WithThreadCount sets the number of threads (concurrent goroutines in execution) that the MapReduce should use.
func WithThreadCount(num int) Option {
	return func(c *config) Option {
		save := c.threadCount

		c.threadCount = num

		return WithThreadCount(save)
	}
}

// WithMapperCount sets the number of Mappers (total number of goroutine tasks) that the MapReduce should use.
func WithMapperCount(num int) Option {
	return func(c *config) Option {
		save := c.mapperCount

		c.mapperCount = num

		return WithMapperCount(save)
	}
}

// WithMaxStripeSize sets the maximum number of elements that each Mapper will receive.
//
// If after calculating the work for each Mapper,
// the number of elements to be handled per Mapper is greater than this value,
// then the number of Mappers will be increased to a point,
// where the number of elements handled by each Mapper is less than this value.
func WithMaxStripeSize(size int) Option {
	return func(c *config) Option {
		save := c.stripeSize

		c.stripeSize = size

		return WithMaxStripeSize(save)
	}
}

// WithOrdering sets the ordering state of the Reduce phase.
//
// When the Reduce phase is ordered, then the Reduce phase of each Mapper will happen in Order,
// that is, the end of the Reduce for range [0,1) HAPPENS BEFORE the start of the Reduce for range [1,2).
//
// Even if all Mappers complete before the first Mapper completes,
// the Reduce for the first Mapper will execute first.
func WithOrdering(ordered bool) Option {
	return func(c *config) Option {
		save := c.ordered

		c.ordered = ordered

		return WithOrdering(save)
	}
}
