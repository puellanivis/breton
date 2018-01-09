// Package sort provides sorting implementations for all builtin data-types and an implementation of Radix sort.
package sort // import "github.com/puellanivis/breton/lib/sort"

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
	case [][]byte:
		sort.Stable(ByteSliceSlice(a))
	case [][]rune:
		sort.Stable(RuneSliceSlice(a))

	default:
		// fallback: use the builtin sort
		sort.Stable(a.(sort.Interface))
	}
}

// we need to wrap these two together to cover all of the operations that need be reversed into one interface.
type reversable interface {
	RadixInterface
	Comparer
}

// now that the functions are all wrapped up together, we can use a single anonymous interface to wrap this.
type reverse struct {
	reversable
}

func (a reverse) Less(i, j int) bool {
	return !a.reversable.Less(i, j)
}

func (a reverse) Compare(i, j int) int {
	return -a.reversable.Compare(i, j)
}

func (a reverse) CompareFunc(x interface{}) func(int) int {
	f := a.reversable.CompareFunc(x)

	return func(i int) int { return -f(i) }
}

func (a reverse) RadixFunc(r int) RadixTest {
	f := a.reversable.RadixFunc(r)

	return func(i int) bool {
		return !f(i)
	}
}

func Reverse(a interface{}) sort.Interface {
	if a == nil {
		return sort.Reverse(sort.Interface(nil))
	}

	switch a := a.(type) {
	case reversable:
		return &reverse{a}

	case []uint:
		return &reverse{UintSlice(a)}
	case []uint8:
		return &reverse{Uint8Slice(a)}
	case []uint16:
		return &reverse{Uint16Slice(a)}
	case []uint32:
		return &reverse{Uint32Slice(a)}
	case []uint64:
		return &reverse{Uint64Slice(a)}

	case []int:
		return &reverse{IntSlice(a)}
	case []int8:
		return &reverse{Int8Slice(a)}
	case []int16:
		return &reverse{Int16Slice(a)}
	case []int32:
		return &reverse{Int32Slice(a)}
	case []int64:
		return &reverse{Int64Slice(a)}

	case []float32:
		return &reverse{Float32Slice(a)}
	case []float64:
		return &reverse{Float64Slice(a)}

	case []string:
		return &reverse{StringSlice(a)}
	case [][]byte:
		return &reverse{ByteSliceSlice(a)}
	case [][]rune:
		return &reverse{RuneSliceSlice(a)}
	}

	return sort.Reverse(a.(sort.Interface))
}

func IsSorted(a interface{}) bool {
	if a == nil {
		return true
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
	case [][]byte:
		return sort.IsSorted(ByteSliceSlice(a))
	case [][]rune:
		return sort.IsSorted(RuneSliceSlice(a))
	}

	return sort.IsSorted(a.(sort.Interface))
}
