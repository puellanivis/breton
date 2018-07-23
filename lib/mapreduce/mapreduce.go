package mapreduce

import (
	"context"
	"reflect"
	"runtime"
)

// DefaultThreadCount defines the number of threads the mapreduce package will assume it should use.
var DefaultThreadCount = runtime.NumCPU()

// A Mapper processes across a set of data.
//
// The mapreduce package will call a Mapper with one of:
//	* a subslice from a slice;
//	* a slice of MapKey values in the type the MapKeys are (i.e. not reflect.Value);
//	* a receive-only channel;
//	* a disjoin sub-range of a Range struct from this package.
//
// Examples:
//	mr.Run(ctx, []string{ ... })          -> mapper.Map(ctx, slice[i:j:j])
//	mr.Run(ctx, map[string]int{})         -> mapper.Map(ctx, []string{ /* subset of key values here */ })
//	mr.Run(ctx, (chan string)(ch))        -> mapper.Map(ctx, (<-chan string)(ch))
//	mr.Run(ctx, Range{Start: 0, End: n})) -> mapper.Map(ctx, Range{Start: i, End: j})
//
// While technically, a Mapper could potentially receive any of these data-types,
// a Mapper SHOULD NOT have to account for all data-types being passed.
// Thereforce, code SHOULD be designed to ensure only one data-type is passed in and out of a Mapper.
//
// As each Map call could be made in parallel,
// a Mapper MUST be thread-safe,
// but SHOULD NOT use any synchronization to ensure thread-safety.
// So, a Mapper SHOULD work on either disjoint data, or read-only access of common data.
//
// A Mapper MUST NOT perform concurrent writes of common data,
// this being the domain of a Reducer.
type Mapper interface {
	Map(ctx context.Context, in interface{}) (out interface{}, err error)
}

// A MapFunc is an adapter to use ordinary functions as a Mapper.
//
// As a repeated note from Mapper documentation,
// this code could run in parallel,
// a MapFunc MUST be thread-safe,
// and SHOULD NOT contain any critical section code.
type MapFunc func(ctx context.Context, in interface{}) (out interface{}, err error)

// Map returns f(ctx, in)
func (f MapFunc) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	return f(ctx, in)
}

// A Reducer processes the results of a Mapper within a critical-section.
//
// The mapreduce package will make a call to Reduce with each of the outputs from a Mapper,
// and ensures that each call to Reduce is in a mutex-locked critical section.
//
// As a Reducer will always be in a critical-section when called from the mapreduce package,
// a Reducer SHOULD NOT be required to perform any synchronization of its own,
// and MAY read and write any common data without concern of another call to a Reducer running in parallel.
type Reducer interface {
	Reduce(ctx context.Context, in interface{}) error
}

// ReduceFunc is an adaptor to use ordinary functions as a Reducer.
//
// As each call to a ReduceFunc from the mapreduce package is called from within a critical section,
// a Reduce MAY read and write any common data without concern of another call to ReduceFunc running in parallel.
type ReduceFunc func(ctx context.Context, in interface{}) error

// Reduce returns r(ctx, in).
func (r ReduceFunc) Reduce(ctx context.Context, in interface{}) error {
	return r(ctx, in)
}

// A MapReduce is a composed pair of a Mapper and a Reducer,
// along with any default Option values that one might wish to setup.
//
// A MapReduce MAY contain only a Mapper, and not a Reducer.
// Such a MapReduce still implements Reducer,
// but will not actually do anything within the Reduce call.
type MapReduce struct {
	m Mapper
	r Reducer

	// shallow copies of this config are made often, do not make this a pointer.
	conf config
}

// New returns a new MapReduce object which defines a whole Mapper/Reducer pair that defines a MapReduce.
// It also can set any Option values that will be the default for any calls to Run.
func New(mapper Mapper, reducer Reducer, opts ...Option) *MapReduce {
	if mapper == nil {
		panic("a MapReduce must have at least a Mapper")
	}

	mr := &MapReduce{
		m: mapper,
		r: reducer,
	}

	for _, opt := range opts {
		_ = opt(&mr.conf)
	}

	return mr
}

// Map invokes the Mapper defined for the MapReduce.
func (mr *MapReduce) Map(ctx context.Context, in interface{}) (interface{}, error) {
	return mr.m.Map(ctx, in)
}

// Reduce invokes the Reducer defined for the MapReduce,
// or simply returns nil if no Reducer was defined.
func (mr *MapReduce) Reduce(ctx context.Context, in interface{}) error {
	if mr.r == nil {
		return nil
	}

	return mr.r.Reduce(ctx, in)
}

// Run performs the MapReduce over the data given, overriding any defaults with the given Options.
// Run returns a receive-only channel of errors that will report all errors returned from a Mapper or Reducer,
// and which is closed upon completion of all Mappers and Reducers.
//
// Run can be called with any of:
//	* a slice or array of any type, where each Mapper will be called with a subslice of the data,
//	* a map of any type, where each Mapper will be called with a slice of a subset of the keys of that map,
//	* a channel of any type, where each Mapper will be called with a receive-only copy of that channel
//	* a Range struct from this package, where each Mapper will receive a disjoint sub-range of that Range.
//
// Any pointer or interface will be dereferenced until Run reaches a concrete type.
// A call to Run that is done on a slice, or map of length 0 (zero), completes immediately with no error.
//
// In order to ensure efficient Mappers, Run SHOULD only ever be called with one type of data.
// In order to process more than one data type, one SHOULD implement two different Mappers.
func (mr *MapReduce) Run(ctx context.Context, data interface{}, opts ...Option) <-chan error {
	v := reflect.ValueOf(data)
	kind := v.Kind()

	for v.IsValid() && (kind == reflect.Ptr || kind == reflect.Interface) {
		if !v.Elem().IsValid() {
			break
		}

		v = v.Elem()
		kind = v.Kind()
		data = v.Interface()
	}

	switch kind {
	case reflect.Chan:
		// No short-circuit check possible.

	case reflect.Slice, reflect.Array, reflect.Map:
		// If it has no elements, short-circuit succeed.
		if v.Len() < 1 {
			ch := make(chan error)
			close(ch)

			return ch
		}

	case reflect.Struct:
		// If we are _not_ a Range, then we‘re a bad type.
		if _, ok := data.(Range); !ok {
			panic("bad type passed to MapReduce.Run")
		}

	default:
		// Anything else is a bad type.
		panic("bad type passed to MapReduce.Run")
	}

	e := &engine{
		m: mr.m,
		r: mr.r,

		conf: mr.conf,
	}

	for _, opt := range opts {
		_ = opt(&e.conf)
	}

	if r, ok := data.(Range); ok {
		// As a Range, we are already setup for the engine.run() call.
		return e.run(ctx, r)
	}

	switch kind {
	case reflect.Chan:
		typ := v.Type()

		switch typ.ChanDir() {
		case reflect.RecvDir:
			// channel is already read-only, we do not need to do anything further here.
		case reflect.BothDir:
			v = v.Convert(reflect.ChanOf(reflect.RecvDir, typ.Elem()))

		default:
			panic("channel as input to mapper must allow receive")
		}

		e.m = MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
			return mr.Map(ctx, v.Interface())
		})

		n := e.conf.threadCount
		if n <= 0 {
			n = DefaultThreadCount

			if n < 1 {
				// Even if the default was set to less than one, we want to ensure it is at least one.
				n = 1
			}

			// Now, make sure that the thread count used in this engine is the same as used here.
			e.conf.threadCount = n
		}

		return e.run(ctx, Range{End: n})

	case reflect.Slice, reflect.Array:
		e.m = MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
			r := in.(Range)

			return mr.Map(ctx, v.Slice3(r.Start, r.End, r.End).Interface())
		})

		return e.run(ctx, Range{End: v.Len()})

	case reflect.Map:
		// We extract and freeze a slice of mapkeys, so that there is a canonical list for all map calls.
		keys := v.MapKeys()

		// get the slice type for []<MapKeyType>
		typ := reflect.SliceOf(v.Type().Key())

		e.m = MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
			r := in.(Range)

			// Here, we build the slice that we will pass in,
			// so that rather than each map call receiving a []reflect.Value, they get a []<MapKeyType>.
			sl := reflect.MakeSlice(typ, 0, r.Width())

			// Since there is non-trivial work necessary to convert the slice types,
			// and we are already splitting the work load through our MapReduce engine,
			// we can do this []reflect.Value -> []<MapKeyType> as a part of the map call process,
			// so that the costs are spread across each map the same as the rest of the mapper work.
			for _, key := range keys[r.Start:r.End] {
				sl = reflect.Append(sl, key)
			}

			return mr.Map(ctx, sl.Interface())
		})

		return e.run(ctx, Range{End: len(keys)})
	}

	// As a final sanity check, we panic with bad type here.
	panic("bad type passed to MapReduce.Run")
}

// Run executes over the given data a new MapReduce constructed from the given Mapper,
// if the given Mapper also implements Reducer,
// then this Reducer is used for the MapReduce.
func Run(ctx context.Context, mapper Mapper, data interface{}, opts ...Option) <-chan error {
	var reducer Reducer

	if r, ok := mapper.(Reducer); ok {
		reducer = r
	}

	return New(mapper, reducer, opts...).Run(ctx, data)
}
