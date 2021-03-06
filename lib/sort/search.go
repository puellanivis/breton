package sort

import (
	"sort"
)

// Comparer defines a set of Comparison functions that will return <0 for i<j, ==0 for i==j, and >0 for i>j.
type Comparer interface {
	Compare(i, j int) int
	CompareFunc(x interface{}) func(int) int
}

// Search returns sort.Search(n, f)
func Search(n int, f func(int) bool) int {
	return sort.Search(n, f)
}

// SearchFor takes arbitrary arguments, and attempts to find x as an element of a.
//
// If a implements `interface{ SearchFor(x interface{}) int }` the results of calling this method are returned.
// If a is a Comparer this is used as the function of sort.Search.
// All basic slices of types supported by this packager are also accepted.
func SearchFor(a interface{}, x interface{}) int {
	type searcherFor interface {
		SearchFor(x interface{}) int
	}

	type comparer interface {
		Len() int
		Comparer
	}

	switch a := a.(type) {
	case searcherFor:
		return a.SearchFor(x)

	case comparer:
		f := a.CompareFunc(x)
		return sort.Search(a.Len(), func(i int) bool { return f(i) >= 0 })

	case []uint:
		return SearchUints(a, x.(uint))
	case []uint8:
		return SearchUint8s(a, x.(uint8))
	case []uint16:
		return SearchUint16s(a, x.(uint16))
	case []uint32:
		return SearchUint32s(a, x.(uint32))
	case []uint64:
		return SearchUint64s(a, x.(uint64))

	case []int:
		return SearchInts(a, x.(int))
	case []int8:
		return SearchInt8s(a, x.(int8))
	case []int16:
		return SearchInt16s(a, x.(int16))
	case []int32:
		return SearchInt32s(a, x.(int32))
	case []int64:
		return SearchInt64s(a, x.(int64))

	case []float32:
		return SearchFloat32s(a, x.(float32))
	case []float64:
		return SearchFloat64s(a, x.(float64))

	case []string:
		return SearchStrings(a, x.(string))
	case [][]byte:
		return SearchByteSlices(a, x.([]byte))
	case [][]rune:
		return SearchRuneSlices(a, x.([]rune))
	}

	panic("sort.Search passed an unknown type")
}
