package sort

import (
	"sort"
)

// IntSlice attaches the methods of sort.Interface to []int, sorting in increasing order.
type IntSlice []int

func (p IntSlice) Len() int           { return len(p) }
func (p IntSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p IntSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p IntSlice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if v < 0 {
			return 0, uintMSB
		}

		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return uintMSB - int(r), uintMSB
}

func (p IntSlice) RadixFunc(r int) RadixTest {
	if r == 0 {
		return func(i int) bool {
			return p[i] >= 0
		}
	}
	mask := int(1) << uint(uintMSB-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p IntSlice) Sort()            { radix(p) }
func (p IntSlice) Search(x int) int { return SearchInts(p, x) }
func (p IntSlice) Radix()           { radix(p) }

// Ints sorts a slice of ints in increasing order.
func Ints(a []int) { radix(IntSlice(a)) }

//SearchInts searches for x in a sorted slice of ints and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInts(a []int, x int) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// IntsAreSorted tests whether a slice of ints is sorted in increasing order.
func IntsAreSorted(a []int) bool { return sort.IsSorted(IntSlice(a)) }

// IntSlice attaches the methods of sort.Interface to []int, sorting in increasing order.
type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Int64Slice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if v < 0 {
			return 0, 63
		}
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return 63 - int(r), 63
}
func (p Int64Slice) RadixFunc(r int) RadixTest {
	if r == 0 {
		return func(i int) bool {
			return p[i] >= 0
		}
	}

	mask := int64(1) << uint(63-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Int64Slice) Sort()              { radix(p) }
func (p Int64Slice) Search(x int64) int { return SearchInt64s(p, x) }
func (p Int64Slice) Radix()             { radix(p) }

// Int64s sorts a slice of int64s in increasing order.
func Int64s(a []int64) { radix(Int64Slice(a)) }

//SearchInt64s searches for x in a sorted slice of int64s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt64s(a []int64, x int64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int64sAreSorted tests whether a slice of int64s is sorted in increasing order.
func Int64sAreSorted(a []int64) bool { return sort.IsSorted(Int64Slice(a)) }

// Int32Slice attaches the methods of sort.Interface to []int32, sorting in increasing order.
type Int32Slice []int32

func (p Int32Slice) Len() int           { return len(p) }
func (p Int32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Int32Slice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if v < 0 {
			return 0, 31
		}
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return 31 - int(r), 31
}
func (p Int32Slice) RadixFunc(r int) RadixTest {
	if r == 0 {
		return func(i int) bool {
			return p[i] >= 0
		}
	}

	mask := int32(1) << uint(31-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Int32Slice) Sort()              { radix(p) }
func (p Int32Slice) Search(x int32) int { return SearchInt32s(p, x) }
func (p Int32Slice) Radix()             { radix(p) }

// Int32s sorts a slice of int32s in increasing order.
func Int32s(a []int32) { radix(Int32Slice(a)) }

//SearchInt32s searches for x in a sorted slice of int32s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt32s(a []int32, x int32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int32sAreSorted tests whether a slice of int32s is sorted in increasing order.
func Int32sAreSorted(a []int32) bool { return sort.IsSorted(Int32Slice(a)) }

// Int16Slice attaches the methods of sort.Interface to []int16, sorting in increasing order.
type Int16Slice []int16

func (p Int16Slice) Len() int           { return len(p) }
func (p Int16Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int16Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Int16Slice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if v < 0 {
			return 0, 15
		}
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return 15 - int(r), 15
}
func (p Int16Slice) RadixFunc(r int) RadixTest {
	if r == 0 {
		return func(i int) bool {
			return p[i] >= 0
		}
	}

	mask := int16(1) << uint(15-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Int16Slice) Sort()              { radix(p) }
func (p Int16Slice) Search(x int16) int { return SearchInt16s(p, x) }
func (p Int16Slice) Radix()             { radix(p) }

// Int16s sorts a slice of int16s in increasing order.
func Int16s(a []int16) { radix(Int16Slice(a)) }

//SearchInt16s searches for x in a sorted slice of int16s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt16s(a []int16, x int16) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int16sAreSorted tests whether a slice of int16s is sorted in increasing order.
func Int16sAreSorted(a []int16) bool { return sort.IsSorted(Int16Slice(a)) }

// Int8Slice attaches the methods of sort.Interface to []int8, sorting in increasing order.
type Int8Slice []int8

func (p Int8Slice) Len() int           { return len(p) }
func (p Int8Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int8Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Int8Slice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if v < 0 {
			return 0, 7
		}
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return 7 - int(r), 7
}
func (p Int8Slice) RadixFunc(r int) RadixTest {
	if r == 0 {
		return func(i int) bool {
			return p[i] >= 0
		}
	}

	mask := int8(1) << uint(7-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Int8Slice) Sort()             { radix(p) }
func (p Int8Slice) Search(x int8) int { return SearchInt8s(p, x) }
func (p Int8Slice) Radix()            { radix(p) }

// Int8s sorts a slice of int8s in increasing order.
func Int8s(a []int8) { radix(Int8Slice(a)) }

//SearchInt8s searches for x in a sorted slice of int8s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt8s(a []int8, x int8) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int8sAreSorted tests whether a slice of int8s is sorted in increasing order.
func Int8sAreSorted(a []int8) bool { return sort.IsSorted(Int8Slice(a)) }
