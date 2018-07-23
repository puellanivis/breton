package sort

import (
	"sort"
	_ "unsafe" // this is to explicitly signal this file is unsafe.
)

//go:linkname quickSort sort.quickSort
func quickSort(data sort.Interface, a, b, maxDepth int)

//go:linkname maxDepth sort.maxDepth
func maxDepth(n int) int
