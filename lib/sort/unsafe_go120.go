package sort

import (
	"math/bits"
	"sort"
	_ "unsafe" // this is to explicitly signal this file is unsafe.
)

//go:linkname heapSort sort.heapSort
func heapSort(data sort.Interface, a, b int)

// quickSort is just an aliased call into pdqsort,
// this means sort.go doesn’t need to be different between go1.19 and earlier.
func quickSort(data sort.Interface, a, b, _ int) {
	heapSort(data, a, b)
}

// maxDepth returns a threashold at which quicksort should switch to heapsort.
// It returns 2 × ceil( log₂(n+1) )
func maxDepth(n int) int {
	if n <= 0 {
		return 0
	}
	return 2 * bits.Len(uint(n))
}
