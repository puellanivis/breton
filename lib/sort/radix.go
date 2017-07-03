package sort

import (
	"sort"
)

// Returns true if the i'th element of the sort.RadixInterface is set
type RadixTest func(i int) bool

type RadixInterface interface {
	Interface

	// Returns start, and end of values to run through for RadixFunc
	RadixRange() (int, int)
	RadixFunc(r int) RadixTest
}

func Radix(a interface{}) {
	if a == nil {
		return
	}

	if a, ok := a.(RadixInterface); ok {
		radix(a)
		return
	}

	switch a := a.(type) {
	case []uint:
		radix(UintSlice(a))
	case []uint8:
		radix(Uint8Slice(a))
	case []uint16:
		radix(Uint16Slice(a))
	case []uint32:
		radix(Uint32Slice(a))
	case []uint64:
		radix(Uint64Slice(a))

	case []int:
		radix(IntSlice(a))
	case []int8:
		radix(Int8Slice(a))
	case []int16:
		radix(Int16Slice(a))
	case []int32:
		radix(Int32Slice(a))
	case []int64:
		radix(Int64Slice(a))

	case []float64:
		radix(Float64Slice(a))
	case []float32:
		radix(Float32Slice(a))

	case []string:
		radix(StringSlice(a))

	default:
		// none of these fit, but if it implements sort.Interface
		// then we can just use the built in sort.Sort anyways.
		if a, ok := a.(sort.Interface); ok {
			sort.Sort(a)
			return
		}

		panic("sort.Radix was passed an unknown type")
	}
}

func radix(a RadixInterface) {
	s, e := a.RadixRange()
	radixSort(a, 0, a.Len(), s, e)
}

func radixSort(a RadixInterface, start, end, radix, last int) {
	i := start
	j := end - 1

	if radix > last || i >= j {
		return
	}

	f := a.RadixFunc(radix)

	for i < j {
		// from the start, find the i-th item that satisfies radix.
		for i < j && !f(i) {
			i++
		}

		// from the end, find the j-th item that doesn’t satisfy radix.
		for i < j && f(j) {
			j--
		}

		if j < i {
			// avoid swapping if they’ve passed each other…
			// really no big deal if they’re ==, but *shrug*
			// already doing the test anyways.
			break
		}

		a.Swap(i, j)
	}

	// if the i-th element doesn’t satisfy radix, then
	// we need to increment it so that i == len(head)
	// where head is the slice of items not satisfying radix.
	if !f(i) {
		i++
	}

	// if the j-th element doesn’t satisfy radix, then
	// we need to increment it so that j is the start of tail,
	// where tail is the slice of items satisfying radix.
	if !f(j) {
		j++
	}

	radix++

	radixSort(a, start, i, radix, last)
	radixSort(a, j, end, radix, last)
}

func bsr(u uint64) uint64
