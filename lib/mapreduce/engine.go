package mapreduce

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type mapper interface {
	Map(ctx context.Context, in interface{}) (out interface{}, err error)
}

type reducer interface {
	Reduce(ctx context.Context, in interface{}) error
}

func engine(ctx context.Context, m mapper, r reducer, rng Range) <-chan error {
	width := rng.Width()

	if width <= 0 {
		errch := make(chan error, 1)

		if width < 0 {
			errch <- errors.New("bad range")
		}

		close(errch)
		return errch
	}

	errch := make(chan error)

	mappers := DefaultThreadCount

	stripe := width / mappers
	if width%mappers > 0 {
		stripe++
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(mappers)

	last := rng.Start

	for i := 0; i < mappers; i++ {
		if last >= rng.End {
			wg.Done()
			continue
		}

		s := last
		e := s + stripe

		if e > rng.End {
			e = rng.End
		}

		go func() {
			defer func() {
				wg.Done()
			}()

			rng := Range{
				Start: s,
				End: e,
			}

			out, err := m.Map(ctx, rng)
			if err != nil {
				errch <- err
			}

			if out == nil || r == nil {
				return
			}

			mu.Lock()
			defer mu.Unlock()

			if err := r.Reduce(ctx, out); err != nil {
				errch <- err
			}
		}()

		last = e
	}

	if last != rng.End {
		panic(fmt.Sprintf("dropped entries! %d != %d", last, rng.End))
	}

	go func() {
		defer close(errch)

		wg.Wait()
	}()

	return errch
}

func Run(ctx context.Context, mr MapReducer, data interface{}) <-chan error {
	if r, ok := data.(Range); ok {
		return engine(ctx, mr, mr, r)
	}

	v := reflect.ValueOf(data)

	switch v.Kind() {
	case reflect.Chan:
		m := func(ctx context.Context, in interface{}) (out interface{}, err error) {
			return mr.Map(ctx, v.Interface())
		}

		return engine(ctx, MapFunc(m), mr, Range{ End: DefaultThreadCount })

	case reflect.Slice:
		m := func(ctx context.Context, in interface{}) (out interface{}, err error) {
			r := in.(Range)

			return mr.Map(ctx, v.Slice(r.Start, r.End).Interface())
		}

		return engine(ctx, MapFunc(m), mr, Range{ End: v.Len() })

	case reflect.Map:
		typ := v.Type().Key()
		keys := v.MapKeys()

		m := func(ctx context.Context, in interface{}) (out interface{}, err error) {
			r := in.(Range)

			sl := reflect.MakeSlice(reflect.SliceOf(typ), 0, r.Width())

			for _, key := range keys[r.Start:r.End] {
				sl = reflect.Append(sl, key)
			}

			return mr.Map(ctx, sl.Interface())
		}

		return engine(ctx, MapFunc(m), mr, Range{ End: len(keys) })
	}

	panic("bad type passed to mapreduce.Run")
}
