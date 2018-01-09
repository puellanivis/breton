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
	}

	if SearchInt64s(l, 42) != 4 {
		t.Error("binary search failed for int64 list")
	}
}

func TestFloat32s(t *testing.T) {
	qsortInstead = qsortNever

	l := []float32{42, 5, 7, 2, 3}

	if Float32sAreSorted(l) {
		t.Error("unsorted int64 list reports as sorted")
	}

	Float32s(l)

	if !Float32sAreSorted(l) {
		t.Error("after sorting int64 list reports as not sorted")
	}

	if SearchFloat32s(l, 42) != 4 {
		t.Error("binary search failed for int64 list")
	}
}

var benchIntArray []int
var benchIntOnce sync.Once

const (
	//benchSize = 1000
	benchSize = 10000000
)

func initIntBench() {
	// bit width here is set to ceil(log_2(benchSize)) - 3
	// if we use a larger bit width, then there will be more recursions based on radix
	// than there would be for quicksort (log n).
	// So, our implementation will often switch to using just quickSort.
	// If we allowed for not-in-place sorting, we could use a better log base.
	log := (maxDepth(benchSize) / 2) - 3
	width := log/8 + 1
	mask := byte(0xff) >> (uint(8 - (log % 8)))
	if width < 2 {
		width = 2
	}
	fmt.Printf("using %d bits << %d\n", log, uintMSB-log)
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

		benchIntArray = append(benchIntArray, (val<<uint(uintMSB-log))|0x10)
	}
}

func BenchmarkIntRadixSort(b *testing.B) {
	qsortInstead = qsortSometimes

	b.StopTimer()
	benchIntOnce.Do(initIntBench)

	buf := make([]int, len(benchIntArray))

	for i := 0; i < b.N; i++ {
		copy(buf, benchIntArray)

		b.StartTimer()
		Ints(buf)
		b.StopTimer()

		if !IntsAreSorted(buf) {
			b.Fatal("sorting failed!")
		}
	}
}

func BenchmarkIntOriginalSort(b *testing.B) {
	b.StopTimer()
	benchIntOnce.Do(initIntBench)

	buf := make([]int, len(benchIntArray))

	for i := 0; i < b.N; i++ {
		copy(buf, benchIntArray)

		b.StartTimer()
		sort.Sort(IntSlice(buf))
		b.StopTimer()

		if !IntsAreSorted(buf) {
			b.Fatal("sorting failed!")
		}
	}
}
