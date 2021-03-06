package sort

import (
	"sort"
	_ "unsafe" // this is to explicitly signal this file is unsafe.
)

// RadixTest defines a function that Returns true if the i'th element of the sort.RadixInterface is set
type RadixTest func(i int) bool

// RadixInterface defines the functions necessary for Radix to use a radix sort rather than sort.Sort.
type RadixInterface interface {
	Interface

	// Returns start, and end of radix values to run through for RadixFunc
	RadixRange() (int, int)
	RadixFunc(r int) RadixTest
}

// Radix attempts to perform a radix sort on the given argument.
//
// If the argument does not implement RadixInterface, or is not a builtin basic slice,
// but it implements sort.Interface, then sort.Sort will be called on the argument.
func Radix(a interface{}) {
	if a == nil {
		return
	}

	switch a := a.(type) {
	case RadixInterface:
		radix(a)

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
		// fallback: use the builtin sort.Sort
		sort.Sort(a.(sort.Interface))
	}
}

func radix(a RadixInterface) {
	s, e := a.RadixRange()
	quickRadix(a, 0, a.Len(), s, e+1)
}

func sortTwo(a RadixInterface, i int) {
	if a.Less(i+1, i) {
		a.Swap(i, i+1)
	}
}

func quickRadix(a RadixInterface, start, end, radix, last int) {
	r := end - start
	if r < 3 {
		if r == 2 {
			sortTwo(a, start)
		}
		return
	}

	if qsortInstead(r, radix, last) {
		quickSort(a, start, end, maxDepth(r))
		return
	}

	radixSort(a, start, end, radix, last)
}

type swapFunc func(i, j int)

func radixPass(f RadixTest, swap swapFunc, start, end int) (pivot int) {
	i, j := start, end-1

	for i < j {
		// from the start, find the i-th item that satisfies radix.
		for i < j && !f(i) {
			i++
		}

		// from the end, find the j-th item that doesn’t satisfy radix.
		for i < j && f(j) {
			j--
		}

		if j <= i {
			// avoid swapping if they’ve passed each other, or are the same thing…
			// while the swap is no big deal, the extra increments not good
			break
		}

		swap(i, j)
		i++
		j--
	}

	// we’re standing on a pivot, if it doesn’t satisfy pivot, then pivot just after.
	if i == j && !f(i) {
		i++
	}

	return i
}

func radixSort(a RadixInterface, start, end, radix, last int) {
	for radix < last {
		r := end - start
		if r < 3 {
			if r == 2 {
				sortTwo(a, start)
			}
			return
		}

		if qsortInstead(r, radix, last) {
			quickSort(a, start, end, maxDepth(r))
			return
		}

		pivot := radixPass(a.RadixFunc(radix), a.Swap, start, end)

		radix++

		if pivot-start < end-pivot {
			quickRadix(a, start, pivot, radix, last)
			start = pivot

		} else {
			quickRadix(a, pivot, end, radix, last)
			end = pivot
		}
	}
}
