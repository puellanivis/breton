package mapreduce

type config struct {
	threadCount int
	mapperCount int

	// We can only have one bound on stripe size: maximum or minimum.
	// So we encode them to the same integer as such:
	// * If zero, there is no minimum or maximum limit,
	// * If negative, the magnitude is the minimum stripe size to be used.
	// * If positive, the magnitude is the maximum stripe size to be used.
	stripeSize int

	ordered bool
}

// Option defines a function that applies a setting or value to a MapReduce configuration.
type Option func(mr *MapReduce) Option

// WithThreadCount sets the number of threads (concurrently executing goroutines) that the MapReduce should use.
func WithThreadCount(num int) Option {
	return func(mr *MapReduce) Option {
		mr.mu.Lock()
		defer mr.mu.Unlock()

		save := mr.conf.threadCount

		mr.conf.threadCount = num

		return WithThreadCount(save)
	}
}

// WithMapperCount sets the number of Mappers (total number of goroutine tasks) that the MapReduce should use.
//
// This value MAY BE overridden, if stripe size values are also specified.
func WithMapperCount(num int) Option {
	return func(mr *MapReduce) Option {
		mr.mu.Lock()
		defer mr.mu.Unlock()

		save := mr.conf.mapperCount

		mr.conf.mapperCount = num

		return WithMapperCount(save)
	}
}

func withStripeSize(size int) Option {
	return func(mr *MapReduce) Option {
		mr.mu.Lock()
		defer mr.mu.Unlock()

		save := mr.conf.stripeSize

		// We encode a minimum limit as a negative value.
		mr.conf.stripeSize = size

		return withStripeSize(save)
	}
}

// WithMinStripeSize sets the minimum number of elements that each Mapper will receive.
//
// One cannot set both a minimum and maximum stripe size setting,
// otherwise we could end up in a case where neither constraint could be met.
// Therefore, setting a minimum stripe size will unset any maximum stripe size setting.
//
// If after calculating the work for each Mapper,
// the number of elements to be handled per Mapper is less than this value,
// then the number of Mappers will be set to a value,
// where the number of elements handled by each Mapper is greater than or equal to this value.
//
// A value less than 1 resets all stripe size settings, both minimum and maximum.
func WithMinStripeSize(size int) Option {
	if size < 1 {
		size = 0
	}

	return withStripeSize(-size)
}

// WithMaxStripeSize sets the maximum number of elements that each Mapper will receive.
//
// One cannot set both a minimum and maximum stripe size setting,
// otherwise we could end up in a case where neither constraint could be met.
// Therefore, setting a maximum stripe size will unset any minimum stripe size setting.
//
// If after calculating the work for each Mapper,
// the number of elements to be handled per Mapper is greater than this value,
// then the number of Mappers will be set to a value,
// where the number of elements handled by each Mapper is less than or equal to this value.
//
// A value less than 1 resets all stripe size settings, both minimum and maximum.
func WithMaxStripeSize(size int) Option {
	if size < 1 {
		size = 0
	}

	return withStripeSize(size)
}

// WithOrdering sets the ordering state of the Reduce phase.
//
// When the Reduce phase is ordered,
// then the Reduce phase of each Mapper will happen in Order,
// that is, given a < b < c,
// the end of the Reduce for range [a,b) HAPPENS BEFORE the start of the Reduce for range [b,c).
//
// So, when the Reduce phase is ordered,
// even if all other Mappers complete before the first Mapper completes,
// the Reduce for the first Mapper WILL execute first.
func WithOrdering(ordered bool) Option {
	return func(mr *MapReduce) Option {
		mr.mu.Lock()
		defer mr.mu.Unlock()

		save := mr.conf.ordered

		mr.conf.ordered = ordered

		return WithOrdering(save)
	}
}
