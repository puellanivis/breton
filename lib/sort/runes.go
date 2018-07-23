package sort

import (
	"math/bits"
	"sort"
	"unicode"
)

// RuneSliceSlice attaches the methods of sort.Interface to [][]rune, sorting in increasing order.
type RuneSliceSlice [][]rune

// Len implements sort.Interface.
func (p RuneSliceSlice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p RuneSliceSlice) Less(i, j int) bool { return p.cmp(p[i], p[j]) < 0 }

// Swap implements sort.Interface.
func (p RuneSliceSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p RuneSliceSlice) cmp(x, y []rune) int {
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

// Compare implements Comparer.
func (p RuneSliceSlice) Compare(i, j int) int {
	return p.cmp(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p RuneSliceSlice) CompareFunc(x interface{}) func(int) int {
	e := x.([]rune)
	return func(i int) int {
		return p.cmp(p[i], e)
	}
}

var runeMSB = int(31 - bits.LeadingZeros32(uint32(unicode.MaxRune)))

// RadixRange implements RadixInterface.
func (p RuneSliceSlice) RadixRange() (int, int) {
	r := 0
	for _, s := range p {
		if len(s) > r {
			r = len(s)
		}
	}
	return 0, r * runeMSB
}

// RadixFunc implements RadixInterface.
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
func (p RuneSliceSlice) Sort() { radix(p) }

// Search is a convenience method.
func (p RuneSliceSlice) Search(x []rune) int { return SearchRuneSlices(p, x) }

// Radix is a convenience method.
func (p RuneSliceSlice) Radix() { radix(p) }

// RuneSlices sorts a slice of []runes in increasing order.
func RuneSlices(a [][]rune) { radix(RuneSliceSlice(a)) }

//SearchRuneSlices searches for x in a sorted slice of []runes and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchRuneSlices(a [][]rune, x []rune) int {
	p := RuneSliceSlice(a)

	return sort.Search(len(a), func(i int) bool { return p.cmp(a[i], x) >= 0 })
}

// RuneSlicesAreSorted tests whether a slice of []runes is sorted in increasing order.
func RuneSlicesAreSorted(a [][]rune) bool { return sort.IsSorted(RuneSliceSlice(a)) }
