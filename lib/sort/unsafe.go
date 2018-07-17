package sort

import (
	"sort"
	_ "unsafe" // this is to ensure that cgo runs.
)

//go:linkname quickSort sort.quickSort
func quickSort(data sort.Interface, a, b, maxDepth int)

//go:linkname maxDepth sort.maxDepth
func maxDepth(n int) int
