package mapreduce

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
)

type StringCollector struct {
	a [][]string
}

func (sc *StringCollector) Reduce(ctx context.Context, in interface{}) error {
	a := in.([]string)

	sc.a = append(sc.a, a)

	return nil
}

func (sc *StringCollector) reset() {
	sc.a = nil
}

func (sc *StringCollector) collect() []rune {
	var r []rune

	for _, a := range sc.a {
		for _, s := range a {
			r = append(r, []rune(s)...)
		}
	}

	return r
}

func (sc *StringCollector) String() string {
	return string(sc.collect())
}

func (sc *StringCollector) sortAndCollect() string {
	r := sc.collect()

	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })

	return string(r)
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

func TestUnorderedMapReduceOverSlice(t *testing.T) {
	DefaultThreadCount = -1

	sc := &StringCollector{}
	mr := New(stringReceiver, sc, WithThreadCount(1), WithOrdering(false))
	ctx := context.Background()

	for n := -1; n <= len(testInput)+1; n++ {
		n := n

		t.Run(fmt.Sprint(n), func(t *testing.T) {
			sc.reset()

			for err := range mr.Run(ctx, testInput, WithThreadCount(n), WithMapperCount(n)) {
				t.Errorf("unexpected error: %+v", err)
			}

			if n > 0 && len(sc.a) != n {
				t.Log(n, sc.a)
				t.Errorf("wrong number of mappers ran: got %d, but expected %d", len(sc.a), n)
			}

			if got := sc.sortAndCollect(); got != testString {
				t.Errorf("mapreduce did not process all elements: got %q, but expected %q", got, testString)
			}
		})
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
	mr := New(unorderedStringReceiver, sc, WithThreadCount(maxN), WithMapperCount(maxN), WithOrdering(true))
	ctx := context.Background()

	// Here we are running a precondition test, to make sure that WithOrdering(false) will fail the test.
	// The WithOrdering(false) here should override the default WithOrder(true) set on the mapreduce.New()
	wg.Add(maxN - 1)
	for err := range mr.Run(ctx, testInput, WithOrdering(false)) {
		t.Errorf("%d mappers: %+v", maxN, err)
	}

	t.Log("ordering=false", maxN, sc.a)

	if sc.String() == testString {
		t.Fatalf("testing relies on an unordered MapReduce producing an unordered collection.")
	}
	// This concludes the precondition testing.

	sc.reset()

	wg.Add(maxN - 1)
	for err := range mr.Run(ctx, testInput) {
		t.Errorf("%d mappers: %+v", maxN, err)
	}

	t.Log("ordering=true", maxN, sc.a)

	if got := sc.String(); got != testString {
		t.Fatalf("an ordered MapReduce should produce an ordered collection: got %q, but expected %q", got, testString)
	}
}

func TestMapReduceOverMap(t *testing.T) {
	DefaultThreadCount = -1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sc := new(StringCollector)
	mr := New(stringReceiver, sc, WithThreadCount(1))

	m := make(map[string]int)
	for i, v := range testInput {
		m[v] = i
	}

	for n := -1; n <= len(testInput)+1; n++ {
		n := n

		t.Run(fmt.Sprint(n), func(t *testing.T) {
			sc.reset()

			for err := range mr.Run(ctx, m, WithThreadCount(n), WithMapperCount(n)) {
				t.Errorf("unexpected error: %+v", err)
			}

			if n > 0 && len(sc.a) != n {
				t.Log(n, sc.a)
				t.Errorf("wrong number of mappers ran: got %d, but expected %d", len(sc.a), n)
			}

			if got := sc.sortAndCollect(); got != testString {
				t.Errorf("mapreduce did not process all elements: got %q, but expected %q", got, testString)
			}
		})
	}
}

func TestMapReduceOverChannel(t *testing.T) {
	DefaultThreadCount = -1

	sc := &StringCollector{}
	mr := New(chanReceiver, sc, WithThreadCount(1))
	ctx := context.Background()

	for n := -1; n <= len(testInput)+1; n++ {
		n := n

		t.Run(fmt.Sprint(n), func(t *testing.T) {
			sc.reset()

			ch := make(chan string)

			errch := mr.Run(ctx, ch, WithThreadCount(n), WithMapperCount(n))

			go func() {
				defer close(ch)

				for _, s := range testInput {
					ch <- s
				}
			}()

			for err := range errch {
				t.Errorf("unexpected error: %+v", err)
			}

			if n > 0 && len(sc.a) != n {
				t.Log(n, sc.a)
				t.Errorf("wrong number of mappers ran: got %d, but expected %d", len(sc.a), n)
			}

			if got := sc.sortAndCollect(); got != testString {
				t.Errorf("mapreduce did not process all elements: got %q, but expected %q", got, testString)
			}
		})
	}
}
