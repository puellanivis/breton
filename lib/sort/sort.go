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
		a = sort.Interface(nil)
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
	case [][]byte:
		sort.Stable(ByteSliceSlice(a))
	case [][]rune:
		sort.Stable(RuneSliceSlice(a))

	default:
		panic("sort.Stable passed an unknown type")
	}
}

type reversable interface {
	RadixInterface
	Comparer
}

type reversed struct {
	reversable
}

func (a reversed) Less(i, j int) bool {
	return !a.reversable.Less(i, j)
}

func (a reversed) Compare(i, j int) int {
	return -a.reversable.Compare(i, j)
}

func (a reversed) CompareFunc(x interface{}) func(int) int {
	f := a.reversable.CompareFunc(x)

	return func(i int) int { return -f(i) }
}

func (a reversed) RadixFunc(r int) RadixTest {
	f := a.reversable.RadixFunc(r)

	return func(i int) bool {
		return !f(i)
	}
}

func Reverse(a interface{}) sort.Interface {
	if a, ok := a.(reversable); ok {
		return &reversed{a}
	}

	switch a := a.(type) {
	case []uint:
		return reversed{UintSlice(a)}
	case []uint8:
		return reversed{Uint8Slice(a)}
	case []uint16:
		return reversed{Uint16Slice(a)}
	case []uint32:
		return reversed{Uint32Slice(a)}
	case []uint64:
		return reversed{Uint64Slice(a)}

	case []int:
		return reversed{IntSlice(a)}
	case []int8:
		return reversed{Int8Slice(a)}
	case []int16:
		return reversed{Int16Slice(a)}
	case []int32:
		return reversed{Int32Slice(a)}
	case []int64:
		return reversed{Int64Slice(a)}

	case []float32:
		return reversed{Float32Slice(a)}
	case []float64:
		return reversed{Float64Slice(a)}

	case []string:
		return reversed{StringSlice(a)}
	case [][]byte:
		return reversed{ByteSliceSlice(a)}
	case [][]rune:
		return reversed{RuneSliceSlice(a)}
	}

	return sort.Reverse(a.(sort.Interface))
}

func IsSorted(a interface{}) bool {
	if a == nil {
		a = sort.Interface(nil)
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
	case [][]byte:
		return sort.IsSorted(ByteSliceSlice(a))
	case [][]rune:
		return sort.IsSorted(RuneSliceSlice(a))
	}

	panic("sort.IsSorted passed an unknown type")
}
