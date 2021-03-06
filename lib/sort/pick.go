package sort

type picker func(l, radix, last int) bool

var qsortInstead picker = qsortSometimes

func qsortSometimes(l, radix, last int) bool {
	r := uint(last-radix) + 1

	return r >= uint(uintMSB) || l < (1<<r)
}

func qsortNever(l, radix, last int) bool {
	return false
}
