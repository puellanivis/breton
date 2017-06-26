package sort

import (
	"sort"
)

const uintPrototype = ^uint(0)

var uintMSB = int(bsr(uint64(uintPrototype)))

// UintSlice attaches the methods of sort.Uinterface to []uint, sorting in increasing order.
type UintSlice []uint

func (p UintSlice) Len() int           { return len(p) }
func (p UintSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p UintSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p UintSlice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return uintMSB - int(r), uintMSB
}
func (p UintSlice) RadixFunc(r int) RadixTest {
	mask := uint(1) << uint(uintMSB-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p UintSlice) Sort()             { radix(p) }
func (p UintSlice) Search(x uint) int { return SearchUints(p, x) }
func (p UintSlice) Radix()            { radix(p) }

// Uints sorts a slice of uints in increasing order.
func Uints(a []uint) { radix(UintSlice(a)) }

//SearchUints searches for x in a sorted slice of uints and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUints(a []uint, x uint) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// UintsAreSorted tests whether a slice of uints is sorted in increasing order.
func UintsAreSorted(a []uint) bool { return sort.IsSorted(UintSlice(a)) }

// UintSlice attaches the methods of sort.Uinterface to []uint, sorting in increasing order.
type Uint64Slice []uint64

func (p Uint64Slice) Len() int           { return len(p) }
func (p Uint64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Uint64Slice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return 63 - int(r), 63
}
func (p Uint64Slice) RadixFunc(r int) RadixTest {
	mask := uint64(1) << uint(63-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Uint64Slice) Sort()               { radix(p) }
func (p Uint64Slice) Search(x uint64) int { return SearchUint64s(p, x) }
func (p Uint64Slice) Radix()              { radix(p) }

// Uint64s sorts a slice of uint64s in increasing order.
func Uint64s(a []uint64) { radix(Uint64Slice(a)) }

//SearchUint64s searches for x in a sorted slice of uint64s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint64s(a []uint64, x uint64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint64sAreSorted tests whether a slice of uint64s is sorted in increasing order.
func Uint64sAreSorted(a []uint64) bool { return sort.IsSorted(Uint64Slice(a)) }

// Uint32Slice attaches the methods of sort.Uinterface to []uint32, sorting in increasing order.
type Uint32Slice []uint32

func (p Uint32Slice) Len() int           { return len(p) }
func (p Uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Uint32Slice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return 31 - int(r), 31
}
func (p Uint32Slice) RadixFunc(r int) RadixTest {
	mask := uint32(1) << uint(31-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Uint32Slice) Sort()               { radix(p) }
func (p Uint32Slice) Search(x uint32) int { return SearchUint32s(p, x) }
func (p Uint32Slice) Radix()              { radix(p) }

// Uint32s sorts a slice of uint32s in increasing order.
func Uint32s(a []uint32) { radix(Uint32Slice(a)) }

//SearchUint32s searches for x in a sorted slice of uint32s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint32s(a []uint32, x uint32) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint32sAreSorted tests whether a slice of uint32s is sorted in increasing order.
func Uint32sAreSorted(a []uint32) bool { return sort.IsSorted(Uint32Slice(a)) }

// Uint16Slice attaches the methods of sort.Uinterface to []uint16, sorting in increasing order.
type Uint16Slice []uint16

func (p Uint16Slice) Len() int           { return len(p) }
func (p Uint16Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint16Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Uint16Slice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return 15 - int(r), 15
}
func (p Uint16Slice) RadixFunc(r int) RadixTest {
	mask := uint16(1) << uint(15-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Uint16Slice) Sort()               { radix(p) }
func (p Uint16Slice) Search(x uint16) int { return SearchUint16s(p, x) }
func (p Uint16Slice) Radix()              { radix(p) }

// Uint16s sorts a slice of uint16s in increasing order.
func Uint16s(a []uint16) { radix(Uint16Slice(a)) }

//SearchUint16s searches for x in a sorted slice of uint16s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint16s(a []uint16, x uint16) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint16sAreSorted tests whether a slice of uint16s is sorted in increasing order.
func Uint16sAreSorted(a []uint16) bool { return sort.IsSorted(Uint16Slice(a)) }

// Uint8Slice attaches the methods of sort.Uinterface to []uint8, sorting in increasing order.
type Uint8Slice []uint8

func (p Uint8Slice) Len() int           { return len(p) }
func (p Uint8Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Uint8Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p Uint8Slice) RadixRange() (int, int) {
	var r uint64
	for _, v := range p {
		if b := bsr(uint64(v)); b > r {
			r = b
		}
	}
	return 7 - int(r), 7
}
func (p Uint8Slice) RadixFunc(r int) RadixTest {
	mask := uint8(1) << uint(7-r)
	return func(i int) bool {
		return p[i]&mask != 0
	}
}

// Sort is a convenience method.
func (p Uint8Slice) Sort()              { radix(p) }
func (p Uint8Slice) Search(x uint8) int { return SearchUint8s(p, x) }
func (p Uint8Slice) Radix()             { radix(p) }

// Uint8s sorts a slice of uint8s in increasing order.
func Uint8s(a []uint8) { radix(Uint8Slice(a)) }

//SearchUint8s searches for x in a sorted slice of uint8s and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchUint8s(a []uint8, x uint8) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// Uint8sAreSorted tests whether a slice of uint8s is sorted in increasing order.
func Uint8sAreSorted(a []uint8) bool { return sort.IsSorted(Uint8Slice(a)) }
