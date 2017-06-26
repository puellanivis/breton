package sort // import "lib/sort"

import (
	"sort"
)

// Our Interface is just the built in sort.Interface
type Interface interface {
	sort.Interface
}

func Sort(a interface{}) {
	Radix(a)
}

func Stable(a interface{}) {
	if a == nil {
		return
	}

	if a, ok := a.(sort.Interface); ok {
		sort.Stable(a)
		return
	}

	switch a := a.(type) {
	case []uint:
		sort.Stable(UintSlice(a))
	case []uint8:
		sort.Stable(Uint8Slice(a))
	case []uint16:
		sort.Stable(Uint16Slice(a))
	case []uint32:
		sort.Stable(Uint32Slice(a))
	case []uint64:
		sort.Stable(Uint64Slice(a))

	case []int:
		sort.Stable(IntSlice(a))
	case []int8:
		sort.Stable(Int8Slice(a))
	case []int16:
		sort.Stable(Int16Slice(a))
	case []int32:
		sort.Stable(Int32Slice(a))
	case []int64:
		sort.Stable(Int64Slice(a))

	case []float32:
		sort.Stable(Float32Slice(a))
	case []float64:
		sort.Stable(Float64Slice(a))

	case []string:
		sort.Stable(StringSlice(a))

	default:
		panic("sort.Stable passed an unknown type")
	}
}

func Reverse(a interface{}) sort.Interface {
	if a == nil {
		return nil
	}

	if a, ok := a.(sort.Interface); ok {
		return sort.Reverse(a)
	}

	switch a := a.(type) {
	case []uint:
		return sort.Reverse(UintSlice(a))
	case []uint8:
		return sort.Reverse(Uint8Slice(a))
	case []uint16:
		return sort.Reverse(Uint16Slice(a))
	case []uint32:
		return sort.Reverse(Uint32Slice(a))
	case []uint64:
		return sort.Reverse(Uint64Slice(a))

	case []int:
		return sort.Reverse(IntSlice(a))
	case []int8:
		return sort.Reverse(Int8Slice(a))
	case []int16:
		return sort.Reverse(Int16Slice(a))
	case []int32:
		return sort.Reverse(Int32Slice(a))
	case []int64:
		return sort.Reverse(Int64Slice(a))

	case []float32:
		return sort.Reverse(Float32Slice(a))
	case []float64:
		return sort.Reverse(Float64Slice(a))

	case []string:
		return sort.Reverse(StringSlice(a))
	}

	panic("sort.Reverse passed an unknown type")
}

func IsSorted(a interface{}) bool {
	if a == nil {
		return false
	}

	if a, ok := a.(sort.Interface); ok {
		return sort.IsSorted(a)
	}

	switch a := a.(type) {
	case []uint:
		return sort.IsSorted(UintSlice(a))
	case []uint8:
		return sort.IsSorted(Uint8Slice(a))
	case []uint16:
		return sort.IsSorted(Uint16Slice(a))
	case []uint32:
		return sort.IsSorted(Uint32Slice(a))
	case []uint64:
		return sort.IsSorted(Uint64Slice(a))

	case []int:
		return sort.IsSorted(IntSlice(a))
	case []int8:
		return sort.IsSorted(Int8Slice(a))
	case []int16:
		return sort.IsSorted(Int16Slice(a))
	case []int32:
		return sort.IsSorted(Int32Slice(a))
	case []int64:
		return sort.IsSorted(Int64Slice(a))

	case []float32:
		return sort.IsSorted(Float32Slice(a))
	case []float64:
		return sort.IsSorted(Float64Slice(a))

	case []string:
		return sort.IsSorted(StringSlice(a))
	}

	panic("sort.IsSorted passed an unknown type")
	return false
}
