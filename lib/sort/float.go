package sort

import (
	"math"
	"sort"
)

// Float64Slice attaches the methods of sort.Interface to []float64, sorting in increasing order.
type Float64Slice []float64

func (p Float64Slice) Len() int                    { return len(p) }
func (p Float64Slice) Less(i, j int) bool          { return p[i] < p[j] || isNaN64(p[i]) && !isNaN64(p[j]) }
func (p Float64Slice) Swap(i, j int)               { p[i], p[j] = p[j], p[i] }
func (p Float64Slice) SearchFor(x interface{}) int { return p.Search(x.(float64)) }

func (p Float64Slice) cmp(x, y float64) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return 1
}
func (p Float64Slice) Compare(i, j int) int {
	return p.cmp(p[i], p[j])
}
func (p Float64Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(float64)

	return func(i int) int {
		return p.cmp(p[i], e)
	}
}

func (p Float64Slice) RadixRange() (int, int) {
	return 0, 63
}
func (p Float64Slice) RadixFunc(r int) RadixTest {
	if r == 0 {
		return func(i int) bool {
			return p[i] >= 0
		}
	}
	mask := uint64(1) << uint(63-r)
	sign := uint64(1) << 63
	return func(i int) bool {
		bits := math.Float64bits(p[i])
		return (bits&mask != 0) != (bits&sign != 0)
	}
}

// Sort is a convenience method.
func (p Float64Slice) Sort()                { radix(p) }
func (p Float64Slice) Search(x float64) int { return SearchFloat64s(p, x) }

func isNaN64(f float64) bool {
	return f != f
}

// SortFloat64s sorts a slice of float64s in increasing order.
func Float64s(a []float64) { radix(Float64Slice(a)) }

//SearchFloat64s searches for x in a sorted slice of float64s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchFloat64s(a []float64, x float64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Float64sAreSorted tests whether a slice of float64s is sorted in increasing order.
func Float64sAreSorted(a []float64) bool { return sort.IsSorted(Float64Slice(a)) }

// Float32Slice attaches the methods of sort.Interface to []float32, sorting in increasing order.
type Float32Slice []float32

func (p Float32Slice) Len() int                    { return len(p) }
func (p Float32Slice) Less(i, j int) bool          { return p[i] < p[j] || isNaN32(p[i]) && !isNaN32(p[j]) }
func (p Float32Slice) Swap(i, j int)               { p[i], p[j] = p[j], p[i] }
func (p Float32Slice) SearchFor(x interface{}) int { return p.Search(x.(float32)) }

func (p Float32Slice) cmp(x, y float32) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return 1
}
func (p Float32Slice) Compare(i, j int) int {
	return p.cmp(p[i], p[j])
}
func (p Float32Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(float32)

	return func(i int) int {
		return p.cmp(p[i], e)
	}
}

func (p Float32Slice) RadixRange() (int, int) {
	return 0, 31
}
func (p Float32Slice) RadixFunc(r int) RadixTest {
	if r == 0 {
		return func(i int) bool {
			return p[i] >= 0
		}
	}
	mask := uint32(1) << uint(31-r)
	return func(i int) bool {
		return math.Float32bits(p[i])&mask != 0
	}
}

// Sort is a convenience method.
func (p Float32Slice) Sort()                { radix(p) }
func (p Float32Slice) Search(x float32) int { return SearchFloat32s(p, x) }

func isNaN32(f float32) bool {
	return f != f
}

// SortFloat32s sorts a slice of float32s in increasing order.
func Float32s(a []float32) { radix(Float32Slice(a)) }

//SearchFloat32s searches for x in a sorted slice of float32s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchFloat32s(a []float32, x float32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Float32sAreSorted tests whether a slice of float32s is sorted in increasing order.
func Float32sAreSorted(a []float32) bool { return sort.IsSorted(Float32Slice(a)) }
