package sort

import (
	"math/bits"
	"sort"
)

// UintSlice attaches the methods of sort.Interface to []uint, sorting in increasing order.
type UintSlice []uint

// Len implements sort.Interface.
func (p UintSlice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p UintSlice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p UintSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpUint(x, y uint) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p UintSlice) Compare(i, j int) int {
	return cmpUint(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p UintSlice) CompareFunc(x interface{}) func(int) int {
	e := x.(uint)
	return func(i int) int {
		return cmpUint(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p UintSlice) RadixRange() (int, int) {
	allBits := uint((1 << uint(uintMSB)) - 1)
	var anyBits uint
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := uintMSB - bits.TrailingZeros(uint(bitMask))

	return bits.LeadingZeros(uint(bitMask)), end
}

// RadixFunc implements RadixInterface.
func (p UintSlice) RadixFunc(r int) RadixTest {
	mask := uint(1) << uint(uintMSB-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p UintSlice) Sort() { radix(p) }

// Radix is a convenience method.
func (p UintSlice) Radix() { radix(p) }

// Search is a convenience method.
func (p UintSlice) Search(x uint) int { return SearchUints(p, x) }

// SearchFor is a convenience method.
func (p UintSlice) SearchFor(x interface{}) int { return SearchUints(p, x.(uint)) }

// Uints sorts a slice of uints in increasing order.
func Uints(a []uint) { radix(UintSlice(a)) }

// SearchUints searches for x in a sorted slice of uints and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUints(a []uint, x uint) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// UintsAreSorted tests whether a slice of uints is sorted in increasing order.
func UintsAreSorted(a []uint) bool { return sort.IsSorted(UintSlice(a)) }

// Uint64Slice attaches the methods of sort.Interface to []uint64, sorting in increasing order.
type Uint64Slice []uint64

// Len implements sort.Interface.
func (p Uint64Slice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p Uint64Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p Uint64Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpUint64(x, y uint64) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p Uint64Slice) Compare(i, j int) int {
	return cmpUint64(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p Uint64Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(uint64)
	return func(i int) int {
		return cmpUint64(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p Uint64Slice) RadixRange() (int, int) {
	allBits := uint64((1 << uint(63)) - 1)
	var anyBits uint64
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := 63 - bits.TrailingZeros64(uint64(bitMask))

	return bits.LeadingZeros64(uint64(bitMask)), end
}

// RadixFunc implements RadixInterface.
func (p Uint64Slice) RadixFunc(r int) RadixTest {
	mask := uint64(1) << uint(63-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Uint64Slice) Sort() { radix(p) }

// Radix is a convenience method.
func (p Uint64Slice) Radix() { radix(p) }

// Search is a convenience method.
func (p Uint64Slice) Search(x uint64) int { return SearchUint64s(p, x) }

// SearchFor is a convenience method.
func (p Uint64Slice) SearchFor(x interface{}) int { return SearchUint64s(p, x.(uint64)) }

// Uint64s sorts a slice of uint64s in increasing order.
func Uint64s(a []uint64) { radix(Uint64Slice(a)) }

// SearchUint64s searches for x in a sorted slice of uint64s and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint64s(a []uint64, x uint64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint64sAreSorted tests whether a slice of uint64s is sorted in increasing order.
func Uint64sAreSorted(a []uint64) bool { return sort.IsSorted(Uint64Slice(a)) }

// Uint32Slice attaches the methods of sort.Interface to []uint32, sorting in increasing order.
type Uint32Slice []uint32

// Len implements sort.Interface.
func (p Uint32Slice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p Uint32Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpUint32(x, y uint32) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p Uint32Slice) Compare(i, j int) int {
	return cmpUint32(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p Uint32Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(uint32)
	return func(i int) int {
		return cmpUint32(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p Uint32Slice) RadixRange() (int, int) {
	allBits := uint32((1 << uint(31)) - 1)
	var anyBits uint32
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := 31 - bits.TrailingZeros32(uint32(bitMask))

	return bits.LeadingZeros32(uint32(bitMask)), end
}

// RadixFunc implements RadixInterface.
func (p Uint32Slice) RadixFunc(r int) RadixTest {
	mask := uint32(1) << uint(31-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Uint32Slice) Sort() { radix(p) }

// Radix is a convenience method.
func (p Uint32Slice) Radix() { radix(p) }

// Search is a convenience method.
func (p Uint32Slice) Search(x uint32) int { return SearchUint32s(p, x) }

// SearchFor is a convenience method.
func (p Uint32Slice) SearchFor(x interface{}) int { return SearchUint32s(p, x.(uint32)) }

// Uint32s sorts a slice of uint32s in increasing order.
func Uint32s(a []uint32) { radix(Uint32Slice(a)) }

// SearchUint32s searches for x in a sorted slice of uint32s and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint32s(a []uint32, x uint32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint32sAreSorted tests whether a slice of uint32s is sorted in increasing order.
func Uint32sAreSorted(a []uint32) bool { return sort.IsSorted(Uint32Slice(a)) }

// Uint16Slice attaches the methods of sort.Interface to []uint16, sorting in increasing order.
type Uint16Slice []uint16

// Len implements sort.Interface.
func (p Uint16Slice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p Uint16Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p Uint16Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpUint16(x, y uint16) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p Uint16Slice) Compare(i, j int) int {
	return cmpUint16(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p Uint16Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(uint16)
	return func(i int) int {
		return cmpUint16(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p Uint16Slice) RadixRange() (int, int) {
	allBits := uint16((1 << uint(15)) - 1)
	var anyBits uint16
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := 15 - bits.TrailingZeros16(uint16(bitMask))

	return bits.LeadingZeros16(uint16(bitMask)), end
}

// RadixFunc implements RadixInterface.
func (p Uint16Slice) RadixFunc(r int) RadixTest {
	mask := uint16(1) << uint(15-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Uint16Slice) Sort() { radix(p) }

// Radix is a convenience method.
func (p Uint16Slice) Radix() { radix(p) }

// Search is a convenience method.
func (p Uint16Slice) Search(x uint16) int { return SearchUint16s(p, x) }

// SearchFor is a convenience method.
func (p Uint16Slice) SearchFor(x interface{}) int { return SearchUint16s(p, x.(uint16)) }

// Uint16s sorts a slice of uint16s in increasing order.
func Uint16s(a []uint16) { radix(Uint16Slice(a)) }

// SearchUint16s searches for x in a sorted slice of uint16s and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint16s(a []uint16, x uint16) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint16sAreSorted tests whether a slice of uint16s is sorted in increasing order.
func Uint16sAreSorted(a []uint16) bool { return sort.IsSorted(Uint16Slice(a)) }

// Uint8Slice attaches the methods of sort.Interface to []uint8, sorting in increasing order.
type Uint8Slice []uint8

// Len implements sort.Interface.
func (p Uint8Slice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p Uint8Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap implements sort.Interface.
func (p Uint8Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func cmpUint8(x, y uint8) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return +1
}

// Compare implements Comparer.
func (p Uint8Slice) Compare(i, j int) int {
	return cmpUint8(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p Uint8Slice) CompareFunc(x interface{}) func(int) int {
	e := x.(uint8)
	return func(i int) int {
		return cmpUint8(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p Uint8Slice) RadixRange() (int, int) {
	allBits := uint8((1 << uint(7)) - 1)
	var anyBits uint8
	for _, v := range p {
		anyBits |= v
		allBits &= v
	}
	bitMask := anyBits &^ allBits

	end := 7 - bits.TrailingZeros8(uint8(bitMask))

	return bits.LeadingZeros8(uint8(bitMask)), end
}

// RadixFunc implements RadixInterface.
func (p Uint8Slice) RadixFunc(r int) RadixTest {
	mask := uint8(1) << uint(7-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Uint8Slice) Sort() { radix(p) }

// Radix is a convenience method.
func (p Uint8Slice) Radix() { radix(p) }

// Search is a convenience method.
func (p Uint8Slice) Search(x uint8) int { return SearchUint8s(p, x) }

// SearchFor is a convenience method.
func (p Uint8Slice) SearchFor(x interface{}) int { return SearchUint8s(p, x.(uint8)) }

// Uint8s sorts a slice of uint8s in increasing order.
func Uint8s(a []uint8) { radix(Uint8Slice(a)) }

// SearchUint8s searches for x in a sorted slice of uint8s and returns the index as specified by sort.Search.
// The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint8s(a []uint8, x uint8) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint8sAreSorted tests whether a slice of uint8s is sorted in increasing order.
func Uint8sAreSorted(a []uint8) bool { return sort.IsSorted(Uint8Slice(a)) }
