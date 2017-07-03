package sort

import (
	"sort"
)

type Searcher interface {
	Interface
	Comparer
}

type Comparer interface {
	Compare(i, j int) int
	CompareFunc(x interface{}) func(int) int
}

func Search(n int, f func(int) bool) int {
	return sort.Search(n, f)
}

func SearchFor(a interface{}, x interface{}) int {
	if a, ok := a.(Searcher); ok {
		f := a.CompareFunc(x)
		return sort.Search(a.Len(), func(i int) bool { return f(i) >= 0 })
	}

	switch a := a.(type) {
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
