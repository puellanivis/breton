package sort

import ()

type picker func(l, radix, last int) bool

var qsortInstead = qsortSometimes

func qsortSometimes(l, radix, last int) bool {
	r := uint(last - radix)

	return r >= uint(uintMSB) || l < (1<<r)
}

func qsortNever(l, radix, last int) bool {
	return false
}
