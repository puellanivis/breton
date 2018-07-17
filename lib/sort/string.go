package sort

import (
	"sort"
	"strings"
)

// StringSlice attaches the methods of sort.Interface to []string, sorting in increasing order.
type StringSlice []string

func (p StringSlice) Len() int           { return len(p) }
func (p StringSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p StringSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func stringCmp(x, y string) int {
	if x == y {
		return 0
	}

	if x < y {
		return -1
	}

	return 1
}

func (p StringSlice) Compare(i, j int) int {
	return stringCmp(p[i], p[j])
}
func (p StringSlice) CompareFunc(x interface{}) func(int) int {
	e := x.(string)
	return func(i int) int {
		return stringCmp(p[i], e)
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
func (p StringSlice) Sort()  { radix(p) }
func (p StringSlice) Radix() { radix(p) }

func (p StringSlice) Search(x string) int         { return SearchStrings(p, x) }
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
