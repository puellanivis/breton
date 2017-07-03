package sort

import (
	"bytes"
	"sort"
	"unicode"
)

// StringSlice attaches the methods of sort.Interface to []string, sorting in increasing order.
type StringSlice []string

func (p StringSlice) Len() int           { return len(p) }
func (p StringSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p StringSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func CompareStrings(x, y string) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return 1
}
func (p StringSlice) Compare(i, j int) int {
	return CompareStrings(p[i], p[j])
}
func (p StringSlice) CompareFunc(x interface{}) func(int) int {
	e := x.(string)
	return func(i int) int {
		return CompareStrings(p[i], e)
	}
}

func (p StringSlice) RadixRange() (int, int) {
	r := 0
	for _, s := range p {
		if len(s) > r {
			r = len(s)
		}
	}
	return 0, r * 8
}
func (p StringSlice) RadixFunc(r int) RadixTest {
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
func (p StringSlice) Sort()               { radix(p) }
func (p StringSlice) Search(x string) int { return SearchStrings(p, x) }
func (p StringSlice) Radix()              { radix(p) }

// SortStrings sorts a slice of strings in increasing order.
func Strings(a []string) { radix(StringSlice(a)) }

//SearchStrings searches for x in a sorted slice of strings and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchStrings(a []string, x string) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= x })
}

// StringsAreSorted tests whether a slice of strings is sorted in increasing order.
func StringsAreSorted(a []string) bool { return sort.IsSorted(StringSlice(a)) }

// ByteSliceSlice attaches the methods of sort.Interface to []string, sorting in increasing order.
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
func (p ByteSliceSlice) Sort()               { radix(p) }
func (p ByteSliceSlice) Search(x []byte) int { return SearchByteSlices(p, x) }
func (p ByteSliceSlice) Radix()              { radix(p) }

// SortByteSlices sorts a slice of strings in increasing order.
func ByteSlices(a [][]byte) { radix(ByteSliceSlice(a)) }

//SearchByteSlices searches for x in a sorted slice of strings and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchByteSlices(a [][]byte, x []byte) int {
	return sort.Search(len(a), func(i int) bool { return bytes.Compare(a[i], x) >= 0 })
}

// ByteSlicesAreSorted tests whether a slice of strings is sorted in increasing order.
func ByteSlicesAreSorted(a [][]byte) bool { return sort.IsSorted(ByteSliceSlice(a)) }

// RuneSliceSlice attaches the methods of sort.Interface to []string, sorting in increasing order.
type RuneSliceSlice [][]rune

func (p RuneSliceSlice) Len() int           { return len(p) }
func (p RuneSliceSlice) Less(i, j int) bool { return CompareRuneSlices(p[i], p[j]) < 0 }
func (p RuneSliceSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func CompareRuneSlices(x, y []rune) int {
	for n, r := range x {
		if n >= len(y) {
			return -1
		}
		if r < y[n] {
			return -1
		}
		if r > y[n] {
			return 1
		}
	}
	return 0
}
func (p RuneSliceSlice) Compare(i, j int) int {
	return CompareRuneSlices(p[i], p[j])
}
func (p RuneSliceSlice) CompareFunc(x interface{}) func(int) int {
	e := x.([]rune)
	return func(i int) int {
		return CompareRuneSlices(p[i], e)
	}
}

var runeMSB = int(bsr(uint64(unicode.MaxRune)))

func (p RuneSliceSlice) RadixRange() (int, int) {
	r := 0
	for _, s := range p {
		if len(s) > r {
			r = len(s)
		}
	}
	return 0, r * runeMSB
}
func (p RuneSliceSlice) RadixFunc(r int) RadixTest {
	n := r / runeMSB
	mask := rune(1) << uint(runeMSB-(r%runeMSB))

	return func(i int) bool {
		if n >= len(p[i]) {
			return false
		}

		return p[i][n]&mask != 0
	}
}

// Sort is a convenience method.
func (p RuneSliceSlice) Sort()               { radix(p) }
func (p RuneSliceSlice) Search(x []rune) int { return SearchRuneSlices(p, x) }
func (p RuneSliceSlice) Radix()              { radix(p) }

// RuneSlices sorts a slice of slice of runes in increasing order.
func RuneSlices(a [][]rune) { radix(RuneSliceSlice(a)) }

// SearchRuneSlices searches for x in a sorted slice of strings and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchRuneSlices(a [][]rune, x []rune) int {
	return sort.Search(len(a), func(i int) bool { return CompareRuneSlices(a[i], x) < 0 })
}

// RuneSlicesAreSorted tests whether a slice of strings is sorted in increasing order.
func RuneSlicesAreSorted(a [][]rune) bool { return sort.IsSorted(RuneSliceSlice(a)) }
