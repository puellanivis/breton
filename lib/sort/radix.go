package sort

import (
	"sort"
)

// Returns true if the i'th element of the sort.RadixInterface is set
type RadixTest func(i int) bool

type RadixInterface interface {
	Interface

	// Returns start, end, and increment
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
		for (i < j) && !f(i) {
			i++
		}
		if (i < j) && f(j) {
			j--
		}

		a.Swap(i, j)
	}

	if !f(i) {
		i++
	}
	if !f(j) {
		j++
	}

	radix++

	radixSort(a, start, i, radix, last)
	radixSort(a, j, end, radix, last)
}

func bsr(u uint64) uint64
