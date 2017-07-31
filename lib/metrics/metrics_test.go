package metrics

import (
	"sync"
	"testing"
)

var (
	cnt = Counter("counter", "test counter")
	cnt2 = ICounter("counter2", "test integer counter")
)

func BenchmarkPrometheusCounterInc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cnt.c.Inc()
	} 
}

func BenchmarkPrometheusCounterIncWithContention(b *testing.B) {
	var wg sync.WaitGroup

	n := 4

	wg.Add(n)

	for j := 0; j < n; j++ {
		go func() {
			defer wg.Done()

			for i := 0; i < b.N; i++ {
				cnt.c.Inc()
			} 
		}()
	}

	wg.Wait()
}

func BenchmarkIntegerCounterInc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cnt2.c.Inc()
	} 
}

func BenchmarkIntegerCounterIncWithContention(b *testing.B) {
	var wg sync.WaitGroup

	n := 4

	wg.Add(n)

	for j := 0; j < n; j++ {
		go func() {
			defer wg.Done()

			for i := 0; i < b.N; i++ {
				cnt2.c.Inc()
			} 
		}()
	}

	wg.Wait()
}
