package mapreduce

import (
	"context"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

type RuneSlice []rune

func (a RuneSlice) Len() int           { return len(a) }
func (a RuneSlice) Less(i, j int) bool { return a[i] < a[j] }
func (a RuneSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type StringCollector struct {
	a [][]string
}

func (sc *StringCollector) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	a := in.([]string)
	var r []string

	for _, s := range a {
		r = append(r, s)
		runtime.Gosched()
	}

	return r, nil
}

func (sc *StringCollector) Reduce(ctx context.Context, in interface{}) error {
	a := in.([]string)

	sc.a = append(sc.a, a)

	return nil
}

type ChanCollector struct {
	a [][]string
}

func (cc *ChanCollector) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	var r []string

	for s := range in.(<-chan string) {
		r = append(r, s)
		runtime.Gosched()
	}

	return r, nil
}

func (cc *ChanCollector) Reduce(ctx context.Context, in interface{}) error {
	a := in.([]string)

	cc.a = append(cc.a, a)

	return nil
}

func TestMapReduce(t *testing.T) {
	s := "abcdefghijklmnopqrstuvwxyz"
	a := strings.Split(s, "")

	sc := &StringCollector{}
	mr := New(sc, sc, WithThreadCount(1))

	ctx := context.Background()

	f := func(n int) {
		sc.a = nil

		for err := range mr.Run(ctx, a, WithThreadCount(n), WithMapperCount(n)) {
			t.Error(err)
		}

		t.Log(n, sc.a)

		var r RuneSlice
		for _, v := range sc.a {
			for _, s := range v {
				r = append(r, []rune(s)...)
			}
		}

		sort.Sort(r)
		t.Logf("mapreduce([]string, %d): %q", n, string(r))

		if !reflect.DeepEqual(s, string(r)) {
			t.Errorf("mapreduce over map with %d mappers did not process all elemnets, expected %q got %q ", n, s, string(r))
		}
	}

	for i := 0; i <= len(a); i++ {
		f(i)
	}

	m := make(map[string]int)
	for i, v := range a {
		m[v] = i
	}

	f = func(n int) {
		sc.a = nil

		for err := range mr.Run(ctx, m, WithThreadCount(n), WithMapperCount(n)) {
			t.Error(err)
		}

		t.Log(n, sc.a)

		var r RuneSlice
		for _, v := range sc.a {
			for _, s := range v {
				r = append(r, []rune(s)...)
			}
		}

		sort.Sort(r)
		t.Logf("mapreduce(map[string]int, %d): %q", n, string(r))

		if !reflect.DeepEqual(s, string(r)) {
			t.Errorf("mapreduce over map with %d mappers did not process all elemnets, expected %q got %q ", n, s, string(r))
		}
	}

	for i := 0; i <= len(a); i++ {
		f(i)
	}

	cc := &ChanCollector{}
	mr = New(cc, cc)

	f = func(n int) {
		cc.a = nil

		ch := make(chan string)

		go func() {
			defer close(ch)

			for _, s := range a {
				ch <- s
			}
		}()

		for err := range mr.Run(ctx, ch, WithThreadCount(n), WithMapperCount(n)) {
			t.Error(err)
		}

		t.Log(n, cc.a)

		var r RuneSlice
		for _, v := range cc.a {
			for _, s := range v {
				r = append(r, []rune(s)...)
			}
		}

		sort.Sort(r)
		t.Logf("mapreduce(chan string, %d): %q", n, string(r))

		if !reflect.DeepEqual(s, string(r)) {
			t.Errorf("mapreduce over map with %d mappers did not process all elemnets, expected %q got %q ", n, s, string(r))
		}
	}

	for i := 0; i <= len(a); i++ {
		f(i)
	}
}
