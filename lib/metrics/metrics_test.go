package metrics

import (
	"sync"
	"testing"
)

var (
	cnt = Counter("counter", "test counter")
)

func BenchmarkPrometheusCounterInc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cnt.c.Inc()
	}
}

func BenchmarkPrometheusCounterIncWithHeavyContention(b *testing.B) {
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
