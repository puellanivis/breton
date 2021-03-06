package sort

import (
	"math/bits"
	"sort"
)

// IntSlice attaches the methods of sort.Interface to []int, sorting in increasing order.
type IntSlice []int

// Len implements sort.Interface.
func (p IntSlice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p IntSlice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p IntSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpInt(x, y int) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p IntSlice) Compare(i, j int) int {
	return cmpInt(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p IntSlice) CompareFunc(x interface{}) func(int) int {
	e := x.(int)
	return func(i int) int {
		return cmpInt(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p IntSlice) RadixRange() (int, int) {
	allBits := int(^1)
	var anyBits int
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := uintMSB - bits.TrailingZeros(uint(bitMask))

	if bitMask < 0 {
		return 0, end
	}

	return bits.LeadingZeros(uint(bitMask)), end
}

// RadixFunc implements RadixInterface.
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
func (p IntSlice) Sort() { radix(p) }

// Radix is a convenience method.
func (p IntSlice) Radix() { radix(p) }

// Search is a convenience method.
func (p IntSlice) Search(x int) int { return SearchInts(p, x) }

// SearchFor is a convenience method.
func (p IntSlice) SearchFor(x interface{}) int { return SearchInts(p, x.(int)) }

// Ints sorts a slice of ints in increasing order.
func Ints(a []int) { radix(IntSlice(a)) }

// SearchInts searches for x in a sorted slice of ints and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInts(a []int, x int) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// IntsAreSorted tests whether a slice of ints is sorted in increasing order.
func IntsAreSorted(a []int) bool { return sort.IsSorted(IntSlice(a)) }

// Int64Slice attaches the methods of sort.Interface to []int64, sorting in increasing order.
type Int64Slice []int64

// Len implements sort.Interface.
func (p Int64Slice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p Int64Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpInt64(x, y int64) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p Int64Slice) Compare(i, j int) int {
	return cmpInt64(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p Int64Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(int64)
	return func(i int) int {
		return cmpInt64(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p Int64Slice) RadixRange() (int, int) {
	allBits := int64(^1)
	var anyBits int64
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := 63 - bits.TrailingZeros64(uint64(bitMask))

	if bitMask < 0 {
		return 0, end
	}

	return bits.LeadingZeros64(uint64(bitMask)), end
}

// RadixFunc implements RadixInterface.
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
func (p Int64Slice) Sort() { radix(p) }

// Radix is a convenience method.
func (p Int64Slice) Radix() { radix(p) }

// Search is a convenience method.
func (p Int64Slice) Search(x int64) int { return SearchInt64s(p, x) }

// SearchFor is a convenience method.
func (p Int64Slice) SearchFor(x interface{}) int { return SearchInt64s(p, x.(int64)) }

// Int64s sorts a slice of int64s in increasing order.
func Int64s(a []int64) { radix(Int64Slice(a)) }

// SearchInt64s searches for x in a sorted slice of int64s and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt64s(a []int64, x int64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int64sAreSorted tests whether a slice of int64s is sorted in increasing order.
func Int64sAreSorted(a []int64) bool { return sort.IsSorted(Int64Slice(a)) }

// Int32Slice attaches the methods of sort.Interface to []int32, sorting in increasing order.
type Int32Slice []int32

// Len implements sort.Interface.
func (p Int32Slice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p Int32Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p Int32Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpInt32(x, y int32) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p Int32Slice) Compare(i, j int) int {
	return cmpInt32(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p Int32Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(int32)
	return func(i int) int {
		return cmpInt32(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p Int32Slice) RadixRange() (int, int) {
	allBits := int32(^1)
	var anyBits int32
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := 31 - bits.TrailingZeros32(uint32(bitMask))

	if bitMask < 0 {
		return 0, end
	}

	return bits.LeadingZeros32(uint32(bitMask)), end
}

// RadixFunc implements RadixInterface.
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
func (p Int32Slice) Sort() { radix(p) }

// Radix is a convenience method.
func (p Int32Slice) Radix() { radix(p) }

// Search is a convenience method.
func (p Int32Slice) Search(x int32) int { return SearchInt32s(p, x) }

// SearchFor is a convenience method.
func (p Int32Slice) SearchFor(x interface{}) int { return SearchInt32s(p, x.(int32)) }

// Int32s sorts a slice of int32s in increasing order.
func Int32s(a []int32) { radix(Int32Slice(a)) }

// SearchInt32s searches for x in a sorted slice of int32s and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt32s(a []int32, x int32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int32sAreSorted tests whether a slice of int32s is sorted in increasing order.
func Int32sAreSorted(a []int32) bool { return sort.IsSorted(Int32Slice(a)) }

// Int16Slice attaches the methods of sort.Interface to []int16, sorting in increasing order.
type Int16Slice []int16

// Len implements sort.Interface.
func (p Int16Slice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p Int16Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p Int16Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpInt16(x, y int16) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p Int16Slice) Compare(i, j int) int {
	return cmpInt16(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p Int16Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(int16)
	return func(i int) int {
		return cmpInt16(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p Int16Slice) RadixRange() (int, int) {
	allBits := int16(^1)
	var anyBits int16
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := 15 - bits.TrailingZeros16(uint16(bitMask))

	if bitMask < 0 {
		return 0, end
	}

	return bits.LeadingZeros16(uint16(bitMask)), end
}

// RadixFunc implements RadixInterface.
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
func (p Int16Slice) Sort() { radix(p) }

// Radix is a convenience method.
func (p Int16Slice) Radix() { radix(p) }

// Search is a convenience method.
func (p Int16Slice) Search(x int16) int { return SearchInt16s(p, x) }

// SearchFor is a convenience method.
func (p Int16Slice) SearchFor(x interface{}) int { return SearchInt16s(p, x.(int16)) }

// Int16s sorts a slice of int16s in increasing order.
func Int16s(a []int16) { radix(Int16Slice(a)) }

// SearchInt16s searches for x in a sorted slice of int16s and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt16s(a []int16, x int16) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int16sAreSorted tests whether a slice of int16s is sorted in increasing order.
func Int16sAreSorted(a []int16) bool { return sort.IsSorted(Int16Slice(a)) }

// Int8Slice attaches the methods of sort.Interface to []int8, sorting in increasing order.
type Int8Slice []int8

// Len implements sort.Interface.
func (p Int8Slice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p Int8Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p Int8Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpInt8(x, y int8) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p Int8Slice) Compare(i, j int) int {
	return cmpInt8(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p Int8Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(int8)
	return func(i int) int {
		return cmpInt8(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p Int8Slice) RadixRange() (int, int) {
	allBits := int8(^1)
	var anyBits int8
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := 7 - bits.TrailingZeros8(uint8(bitMask))

	if bitMask < 0 {
		return 0, end
	}

	return bits.LeadingZeros8(uint8(bitMask)), end
}

// RadixFunc implements RadixInterface.
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
func (p Int8Slice) Sort() { radix(p) }

// Radix is a convenience method.
func (p Int8Slice) Radix() { radix(p) }

// Search is a convenience method.
func (p Int8Slice) Search(x int8) int { return SearchInt8s(p, x) }

// SearchFor is a convenience method.
func (p Int8Slice) SearchFor(x interface{}) int { return SearchInt8s(p, x.(int8)) }

// Int8s sorts a slice of int8s in increasing order.
func Int8s(a []int8) { radix(Int8Slice(a)) }

// SearchInt8s searches for x in a sorted slice of int8s and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchInt8s(a []int8, x int8) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Int8sAreSorted tests whether a slice of int8s is sorted in increasing order.
func Int8sAreSorted(a []int8) bool { return sort.IsSorted(Int8Slice(a)) }
