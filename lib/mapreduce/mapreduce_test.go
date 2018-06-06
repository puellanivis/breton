package mapreduce

import (
	"context"
	"strings"
	"testing"
)

type StringCollector struct{
	a [][]string
}

func (sc *StringCollector) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	a := in.([]string)

	return a, nil
}

func (sc *StringCollector) Reduce(ctx context.Context, in interface{}) error {
	a := in.([]string)

	sc.a = append(sc.a, a)

	return nil
}

type ChanCollector struct{
	a [][]string
}

func (cc *ChanCollector) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	a := make([]string, 0)

	for s := range in.(<-chan string) {
		a = append(a, s)
	}

	return a, nil
}

func (cc *ChanCollector) Reduce(ctx context.Context, in interface{}) error {
	a := in.([]string)

	cc.a = append(cc.a, a)

	return nil
}

func TestMapReduce(t *testing.T) {
	a := strings.Split("abcdefghijklmnopqrstuvwxyz", "")

	sc := &StringCollector{}
	mr := New(sc, sc, WithThreadCount(1))

	ctx := context.Background()

	f := func(n int) {
		sc.a = nil

		for err := range mr.Run(ctx, a, WithThreadCount(n), WithMapperCount(n)) {
			t.Error(err)
		}

		t.Log(n, sc.a)
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
	}

	for i := 0; i <= len(a); i++ {
		f(i)
	}
}
