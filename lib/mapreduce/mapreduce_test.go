package mapreduce

import (
	"context"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
)

type RuneSlice []rune

func (a RuneSlice) Len() int           { return len(a) }
func (a RuneSlice) Less(i, j int) bool { return a[i] < a[j] }
func (a RuneSlice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type StringCollector struct {
	a [][]string
}

func (sc *StringCollector) Reduce(ctx context.Context, in interface{}) error {
	a := in.([]string)

	sc.a = append(sc.a, a)

	return nil
}

var (
	stringReceiver = MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		var r []string

		for _, s := range in.([]string) {
			r = append(r, s)
			runtime.Gosched()
		}

		return r, nil
	})

	chanReceiver = MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		var r []string

		for s := range in.(<-chan string) {
			r = append(r, s)
			runtime.Gosched()
		}

		return r, nil
	})
)

var (
	testString = "abcdefghijklmnopqrstuvwxyz"
	testInput  = strings.Split(testString, "")
)

func TestMapReduceOverSlice(t *testing.T) {
	DefaultThreadCount = -1

	sc := &StringCollector{}
	mr := New(stringReceiver, sc, WithThreadCount(1))
	ctx := context.Background()

	f := func(n int) {
		sc.a = nil

		for err := range mr.Run(ctx, testInput, WithThreadCount(n), WithMapperCount(n), WithOrdering(false)) {
			t.Errorf("%d mappers: %+v", n, err)
		}

		if n > 0 && len(sc.a) != n {
			t.Log(n, sc.a)
			t.Errorf("wrong number of mappers ran, expected %d, but got %d", n, len(sc.a))
		}

		var r RuneSlice
		for _, v := range sc.a {
			for _, s := range v {
				r = append(r, []rune(s)...)
			}
		}

		sort.Sort(r)
		got := string(r)

		if got != testString {
			t.Logf("mapreduce([]string, %d): %q", n, got)
			t.Errorf("mapreduce over map with %d mappers did not process all elements, expected %q got %q ", n, testString, got)
		}
	}

	for i := -1; i <= len(testInput)+1; i++ {
		f(i)
	}
}

func TestOrderedMapReduceOverSlice(t *testing.T) {
	DefaultThreadCount = -1
	maxN := len(testInput)

	var wg sync.WaitGroup

	unorderedStringReceiver := MapFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		var r []string

		flag := true

		for _, s := range in.([]string) {
			r = append(r, s)

			if s == testInput[0] {
				wg.Wait()
				flag = false
			}
		}

		if flag {
			wg.Done()
		}

		return r, nil
	})

	sc := &StringCollector{}
	mr := New(unorderedStringReceiver, sc, WithThreadCount(1), WithOrdering(true))
	ctx := context.Background()

	// the WithOrdering(false) here should override the default WithOrder(true) set on the mapreduce.New()
	wg.Add(maxN - 1)
	for err := range mr.Run(ctx, testInput, WithThreadCount(maxN), WithMapperCount(maxN), WithOrdering(false)) {
		t.Errorf("%d mappers: %+v", maxN, err)
	}

	t.Log(maxN, sc.a)

	var r RuneSlice
	for _, v := range sc.a {
		for _, s := range v {
			r = append(r, []rune(s)...)
		}
	}

	if string(r) == testString {
		t.Fatalf("testing relies upon runtime.Gosched() producing a non-ordered slice collection.")
	}

	sc.a = nil

	wg.Add(maxN - 1)
	for err := range mr.Run(ctx, testInput, WithThreadCount(maxN), WithMapperCount(maxN)) {
		t.Errorf("%d mappers: %+v", maxN, err)
	}

	t.Log(maxN, sc.a)

	r = nil
	for _, v := range sc.a {
		for _, s := range v {
			r = append(r, []rune(s)...)
		}
	}

	got := string(r)
	if got != testString {
		t.Fatalf("an ordered MapReduce should have returned an ordered slice collection, expected %q, got %q", testString, got)
	}
}

func TestMapReduceOverMap(t *testing.T) {
	DefaultThreadCount = -1

	sc := &StringCollector{}
	mr := New(stringReceiver, sc, WithThreadCount(1))
	ctx := context.Background()

	m := make(map[string]int)
	for i, v := range testInput {
		m[v] = i
	}

	f := func(n int) {
		sc.a = nil

		for err := range mr.Run(ctx, m, WithThreadCount(n), WithMapperCount(n)) {
			t.Errorf("%d mappers: %+v", n, err)
		}

		if n > 0 && len(sc.a) != n {
			t.Log(n, sc.a)
			t.Errorf("wrong number of mappers ran, expected %d, but got %d", n, len(sc.a))
		}

		var r RuneSlice
		for _, v := range sc.a {
			for _, s := range v {
				r = append(r, []rune(s)...)
			}
		}

		sort.Sort(r)
		got := string(r)

		if got != testString {
			t.Logf("mapreduce(map[string]int, %d): %q", n, got)
			t.Errorf("mapreduce over map with %d mappers did not process all elements, expected %q got %q ", n, testString, got)
		}
	}

	for i := -1; i <= len(testInput)+1; i++ {
		f(i)
	}
}

func TestMapReduceOverChannel(t *testing.T) {
	DefaultThreadCount = -1

	sc := &StringCollector{}
	mr := New(chanReceiver, sc, WithThreadCount(1))
	ctx := context.Background()

	f := func(n int) {
		sc.a = nil

		ch := make(chan string)

		errch := mr.Run(ctx, ch, WithThreadCount(n), WithMapperCount(n))

		go func() {
			defer close(ch)

			for _, s := range testInput {
				ch <- s
			}
		}()

		for err := range errch {
			t.Errorf("%d mappers: %+v", n, err)
		}

		if n > 0 && len(sc.a) != n {
			t.Log(n, sc.a)
			t.Errorf("wrong number of mappers ran, expected %d, but got %d", n, len(sc.a))
		}

		var r RuneSlice
		for _, v := range sc.a {
			for _, s := range v {
				r = append(r, []rune(s)...)
			}
		}

		sort.Sort(r)
		got := string(r)

		if got != testString {
			t.Logf("mapreduce(chan string, %d): %q", n, got)
			t.Errorf("mapreduce over map with %d mappers did not process all elements, expected %q got %q ", n, testString, got)
		}
	}

	for i := -1; i <= len(testInput)+1; i++ {
		f(i)
	}
}
