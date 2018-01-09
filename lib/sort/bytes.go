package sort

import (
	"bytes"
	"sort"
)

// ByteSliceSlice attaches the methods of sort.Interface to [][]byte, sorting in increasing order.
type ByteSliceSlice [][]byte

func (p ByteSliceSlice) Len() int           { return len(p) }
func (p ByteSliceSlice) Less(i, j int) bool { return bytes.Compare(p[i], p[j]) < 0 }
func (p ByteSliceSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p ByteSliceSlice) Compare(i, j int) int {
	return bytes.Compare(p[i], p[j])
}
func (p ByteSliceSlice) CompareFunc(x interface{}) func(int) int {
	e := x.([]byte)
	return func(i int) int {
		return bytes.Compare(p[i], e)
	}
}

func (p ByteSliceSlice) RadixRange() (int, int) {
	r := 0
	for _, s := range p {
		if len(s) > r {
			r = len(s)
		}
	}
	return 0, r * 8
}
func (p ByteSliceSlice) RadixFunc(r int) RadixTest {
	n := r / 8
	mask := byte(1 << uint(7-(r&0x7)))

	return func(i int) bool {
		if n >= len(p[i]) {
			return false
		}

		return p[i][n]&mask != 0
	}
}

// Sort is a convenience method.
func (p ByteSliceSlice) Sort()  { radix(p) }
func (p ByteSliceSlice) Radix() { radix(p) }

func (p ByteSliceSlice) Search(x []byte) int         { return SearchByteSlices(p, x) }
func (p ByteSliceSlice) SearchFor(x interface{}) int { return SearchByteSlices(p, x.([]byte)) }

// SortByteSlices sorts a slice of []bytes in increasing order.
func ByteSlices(a [][]byte) { radix(ByteSliceSlice(a)) }

//SearchByteSlices searches for x in a sorted slice of []bytes and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchByteSlices(a [][]byte, x []byte) int {
	return sort.Search(len(a), func(i int) bool { return bytes.Compare(a[i], x) >= 0 })
}

// ByteSlicesAreSorted tests whether a slice of []bytes is sorted in increasing order.
func ByteSlicesAreSorted(a [][]byte) bool { return sort.IsSorted(ByteSliceSlice(a)) }
