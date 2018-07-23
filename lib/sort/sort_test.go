package sort

import (
	"crypto/rand"
	"fmt"
	"sort"
	"sync"
	"testing"
)

func TestInt32s(t *testing.T) {
	qsortInstead = qsortNever

	l := []int32{42, 5, 7, 2, 3}

	if Int32sAreSorted(l) {
		t.Error("unsorted int32 list reports as sorted")
	}

	Int32s(l)

	if !Int32sAreSorted(l) {
		t.Error("after sorting int32 list reports as not sorted")
		t.Error("Got:", l)
	}

	if SearchInt32s(l, 42) != 4 {
		t.Error("binary search failed for int32 list")
	}
}

func TestReverseInt32s(t *testing.T) {
	qsortInstead = qsortNever

	a := []int32{42, 5, 7, 2, 3}
	l := Reverse(a)

	if IsSorted(l) {
		t.Error("unsorted reverse int32 list reports as sorted")
	}

	Sort(l)

	if !IsSorted(l) {
		t.Error("after reverse sorting int32 list reports as not sorted")
		t.Error("Got:", l)
	}

	if got := SearchFor(l, int32(42)); got != 0 {
		t.Errorf("binary search failed for reversed int32 list got %d wanted 0", got)
	}
}

func TestInt64s(t *testing.T) {
	qsortInstead = qsortNever

	l := []int64{42, 5, 7, 2, 3}

	if Int64sAreSorted(l) {
		t.Error("unsorted int64 list reports as sorted")
	}

	Sort(l)

	if !Int64sAreSorted(l) {
		t.Error("after sorting int64 list reports as not sorted")
		t.Error("Got:", l)
	}

	if SearchInt64s(l, 42) != 4 {
		t.Error("binary search failed for int64 list")
	}
}

var benchIntArray = make([][]int, 5)
var benchIntOnce = make([]sync.Once, len(benchIntArray))

const (
	//benchSize = 1000
	benchSize = 10000000
)

func initIntBench(n int) {
	// bit width here is set to ceil(log_2(benchSize)) - n
	// With a larger bit width, there will be more recursions based on radi than for quicksort (i.e. log n).
	// So, the “intelligent” Radix implementation will often switch to using sort.qsort.
	//
	// Quick rundown of pure radix, vs “intelligent” Radix, and sort.Sort:
	// 	n < 1:     pure radix sort is slower; “intelligent” Radix will almost always pick sort.qsort.
	//	n > 3:     pure radix sort is faster; “intelligent” Radix will almost never pick sort.qsort.
	// 	n = 2:     all three run in around the same, and picking one or the other is a wash.
	//      otherwise: “intelligent” Radix will either pick sort.qsort too much or not enough,
	//			so, results end up super non-deterministic, but either n = 1 sucks, or n = 3 does.
	log := (maxDepth(benchSize) / 2) - n
	width := log/8 + 1
	mask := byte(0xff) >> (uint(8 - (log % 8)))
	if width < 2 {
		width = 2
	}
	fmt.Printf("using a slice of len %d\n", benchSize)
	fmt.Printf("n = %d is using %d bits << %d\n", n, log, uintMSB-log)
	b := make([]byte, width)

	for i := 0; i < benchSize; i++ {
		_, err := rand.Read(b)
		if err != nil {
			panic(err)
		}

		var val int
		val = int(b[0] & mask)

		for j := 1; j < len(b); j++ {
			val <<= 8
			val |= int(b[j])
		}

		/* if b[0] & 0x80 != 0 {
			val = -val
		} //*/

		benchIntArray[n] = append(benchIntArray[n], (val<<uint(uintMSB-log))|0x10)
	}
}

func benchmarkIntAllRadixSort(b *testing.B, n int) {
	b.StopTimer()
	benchIntOnce[n].Do(func() { initIntBench(n) })

	qsortInstead = qsortNever

	buf := make([]int, len(benchIntArray[n]))

	for i := 0; i < b.N; i++ {
		copy(buf, benchIntArray[n])

		b.StartTimer()
		Ints(buf)
		b.StopTimer()

		if !IntsAreSorted(buf) {
			b.Fatal("sorting failed!")
		}
	}
}

func BenchmarkIntAllRadixSort0(b *testing.B) {
	benchmarkIntAllRadixSort(b, 0)
}
func BenchmarkIntAllRadixSort1(b *testing.B) {
	benchmarkIntAllRadixSort(b, 1)
}
func BenchmarkIntAllRadixSort2(b *testing.B) {
	benchmarkIntAllRadixSort(b, 2)
}
func BenchmarkIntAllRadixSort3(b *testing.B) {
	benchmarkIntAllRadixSort(b, 3)
}
func BenchmarkIntAllRadixSort4(b *testing.B) {
	benchmarkIntAllRadixSort(b, 4)
}

func benchmarkIntRadixSort(b *testing.B, n int) {
	b.StopTimer()
	benchIntOnce[n].Do(func() { initIntBench(n) })

	qsortInstead = qsortSometimes

	buf := make([]int, len(benchIntArray[n]))

	for i := 0; i < b.N; i++ {
		copy(buf, benchIntArray[n])

		b.StartTimer()
		Ints(buf)
		b.StopTimer()

		if !IntsAreSorted(buf) {
			b.Fatal("sorting failed!")
		}
	}
}

func BenchmarkIntRadixSort0(b *testing.B) {
	benchmarkIntRadixSort(b, 0)
}
func BenchmarkIntRadixSort1(b *testing.B) {
	benchmarkIntRadixSort(b, 1)
}
func BenchmarkIntRadixSort2(b *testing.B) {
	benchmarkIntRadixSort(b, 2)
}
func BenchmarkIntRadixSort3(b *testing.B) {
	benchmarkIntRadixSort(b, 3)
}
func BenchmarkIntRadixSort4(b *testing.B) {
	benchmarkIntRadixSort(b, 4)
}

func benchmarkIntOriginalSort(b *testing.B, n int) {
	b.StopTimer()
	benchIntOnce[n].Do(func() { initIntBench(n) })

	buf := make(IntSlice, len(benchIntArray[n]))

	for i := 0; i < b.N; i++ {
		copy(buf, benchIntArray[n])

		b.StartTimer()
		sort.Sort(buf)
		b.StopTimer()

		if !IntsAreSorted(buf) {
			b.Fatal("sorting failed!")
		}
	}
}

func BenchmarkIntOriginalSort0(b *testing.B) {
	benchmarkIntOriginalSort(b, 0)
}
func BenchmarkIntOriginalSort1(b *testing.B) {
	benchmarkIntOriginalSort(b, 1)
}
func BenchmarkIntOriginalSort2(b *testing.B) {
	benchmarkIntOriginalSort(b, 2)
}
func BenchmarkIntOriginalSort3(b *testing.B) {
	benchmarkIntOriginalSort(b, 3)
}
func BenchmarkIntOriginalSort4(b *testing.B) {
	benchmarkIntOriginalSort(b, 4)
}
