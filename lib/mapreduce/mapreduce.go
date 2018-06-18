package mapreduce

import (
	"context"
	"reflect"
	"runtime"
)

var DefaultThreadCount = runtime.NumCPU()

type Mapper interface {
	Map(ctx context.Context, in interface{}) (out interface{}, err error)
}

// Map defines a function to be called on each Stripe of data in a given MapReduce.
// This is not critical section code and each Map will be run in a separate goroutine in parallel to the thread count.
type MapFunc func(ctx context.Context, in interface{}) (out interface{}, err error)

func (m MapFunc) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	return m(ctx, in)
}

type Reducer interface {
	Reduce(ctx context.Context, in interface{}) error
}

// Reduce defines a function that recieves the output of a single Map.
// This is critical section code, and only one Reduce goroutine will ever be running at a time.
type ReduceFunc func(ctx context.Context, in interface{}) error

func (r ReduceFunc) Reduce(ctx context.Context, in interface{}) error {
	return r(ctx, in)
}

type MapReduce struct {
	m Mapper
	r Reducer

	conf config
}

func New(m Mapper, r Reducer, opts ...Option) *MapReduce {
	mr := &MapReduce{
		m: m,
		r: r,
	}

	for _, opt := range opts {
		_ = opt(&mr.conf)
	}

	return mr
}

func (mr *MapReduce) Map(ctx context.Context, in interface{}) (interface{}, error) {
	if mr.m == nil {
		panic("a MapReduce must implement at least a Mapper")
	}

	return mr.m.Map(ctx, in)
}

func (mr *MapReduce) Reduce(ctx context.Context, in interface{}) error {
	if mr.r == nil {
		return nil
	}

	return mr.r.Reduce(ctx, in)
}

func (mr *MapReduce) Run(ctx context.Context, data interface{}, opts ...Option) <-chan error {
	v := reflect.ValueOf(data)
	kind := v.Kind()

	if kind == reflect.Ptr {
		return mr.Run(ctx, v.Elem().Interface(), opts...)
	}

	e := &engine{
		m: mr.m,
		r: mr.r,

		conf: mr.conf,
	}

	for _, opt := range opts {
		_ = opt(&e.conf)
	}

	for kind == reflect.Interface {
		v = v.Elem()
		kind = v.Kind()
	}

	if r, ok := data.(Range); ok {
		return e.run(ctx, r)
	}

	switch kind {
	case reflect.Chan:
		typ := v.Type()

		switch typ.ChanDir() {
		case reflect.RecvDir:
			// do not need to do anything here.
		case reflect.BothDir:
			v = v.Convert(reflect.ChanOf(reflect.RecvDir, typ.Elem()))

		default:
			panic("channel as input to mapper must allow receive")
		}

		e.m = MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
			return mr.Map(ctx, v.Interface())
		})

		n := e.conf.threadCount
		if n < 1 {
			// if no thread count option was set, then go with the default.
			n = DefaultThreadCount

			if n < 1 {
				// if even the default even empty, then make it at least one.
				n = 1
			}
		}

		return e.run(ctx, Range{End: n})

	case reflect.Slice, reflect.Array:
		e.m = MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
			r := in.(Range)

			return mr.Map(ctx, v.Slice3(r.Start, r.End, r.End).Interface())
		})

		return e.run(ctx, Range{End: v.Len()})

	case reflect.Map:
		// We extract and freeze a slice of mapkeys, so that there is a canonical list for all mappers.
		typ := reflect.SliceOf(v.Type().Key())
		keys := v.MapKeys()

		e.m = MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
			r := in.(Range)

			// Here, we build the slice that we will pass in,
			// so that rather than mappers receiving a []reflect.Value, they get a []<MapKeyType>.

			// Since there is non-trivial work necessary to convert the slice types,
			// we do this as a part of the mapping process,
			// so that the costs are spread across each map the same as the rest of the mapper work.
			sl := reflect.MakeSlice(typ, 0, r.Width())
			for _, key := range keys[r.Start:r.End] {
				sl = reflect.Append(sl, key)
			}

			return mr.Map(ctx, sl.Interface())
		})

		return e.run(ctx, Range{End: len(keys)})
	}

	panic("bad type passed to mapreduce.Run")
}
