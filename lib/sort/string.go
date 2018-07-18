package sort

import (
	"sort"
	"strings"
)

// StringSlice attaches the methods of sort.Interface to []string, sorting in increasing order.
type StringSlice []string

// Len implements sort.Interface.
func (p StringSlice) Len() int { return len(p) }

// Less implements sort.Interface.
func (p StringSlice) Less(i, j int) bool { return strings.Compare(p[i], p[j]) < 0 }

// Swap implements sort.Interface.
func (p StringSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Compare implements Comparer.
func (p StringSlice) Compare(i, j int) int {
	return strings.Compare(p[i], p[j])
}

// CompareFunc implements Comparer.
func (p StringSlice) CompareFunc(x interface{}) func(int) int {
	e := x.(string)
	return func(i int) int {
		return strings.Compare(p[i], e)
	}
}

// RadixRange implements RadixInterface.
func (p StringSlice) RadixRange() (int, int) {
	r := 0
	for _, s := range p {
		if len(s) > r {
			r = len(s)
		}
	}
	return 0, r * 8
}

// RadixFunc implements RadixInterface.
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
func (p StringSlice) Sort() { radix(p) }

// Radix is a convenience method.
func (p StringSlice) Radix() { radix(p) }

// Search is a convenience method.
func (p StringSlice) Search(x string) int { return SearchStrings(p, x) }

// SearchFor is a convenience method.
func (p StringSlice) SearchFor(x interface{}) int { return SearchStrings(p, x.(string)) }

// Strings sorts a slice of strings in increasing order.
func Strings(a []string) { radix(StringSlice(a)) }

//SearchStrings searches for x in a sorted slice of strings and returns the index
// as specified by sort.Search.  The return value is the index to insert x if x is not present (it could be len(a)).
// The slice must be sorted in ascending order.
func SearchStrings(a []string, x string) int {
	return sort.Search(len(a), func(i int) bool { return strings.Compare(a[i], x) >= 0 })
}

// StringsAreSorted tests whether a slice of strings is sorted in increasing order.
func StringsAreSorted(a []string) bool { return sort.IsSorted(StringSlice(a)) }
