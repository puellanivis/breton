package sort

import ()

type Searcher interface {
	Search(x interface{}) int
}

func Search(a interface{}, x interface{}) int {
	if a == nil {
		return 0
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
	}

	if a, ok := a.(Searcher); ok {
		return a.Search(x)
	}

	panic("sort.Search passed an unknown type")
	return -1
}
