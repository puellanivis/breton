package sort

import (
	"testing"
)

func TestFloat64s(t *testing.T) {
	qsortInstead = qsortNever

	l := []float64{42.37, 5.3, -7.5, 2, 3, 0.5, -6, 100000, 0}

	if Float64sAreSorted(l) {
		t.Error("unsorted float64 list reports as sorted")
	}

	radix(Float64Slice(l))

	if !Float64sAreSorted(l) {
		t.Error("after sorting float64 list reports as not sorted")
		t.Log("Got:", l)
	}

	if SearchFloat64s(l, 42.37) != 7 {
		t.Error("binary search failed for float64 list")
	}
}

func TestUints(t *testing.T) {
	qsortInstead = qsortNever

	l := []uint{42, 5, 7, 2, 3}

	if UintsAreSorted(l) {
		t.Error("unsorted uint list reports as sorted")
	}

	radix(UintSlice(l))

	if !UintsAreSorted(l) {
		t.Error("after sorting uint list reports as not sorted")
		t.Log(l)
	}

	if SearchUints(l, 42) != 4 {
		t.Error("binary search failed for uint list")
	}
}

func TestInts(t *testing.T) {
	qsortInstead = qsortNever

	l := []int{42, 5, -7, -2, 3}

	if IntsAreSorted(l) {
		t.Error("unsorted int list reports as sorted")
	}

	radix(IntSlice(l))

	if !IntsAreSorted(l) {
		t.Error("after sorting int list reports as not sorted")
		t.Log(l)
	}

	if SearchInts(l, 42) != 4 {
		t.Error("binary search failed for int list")
	}
}

func TestStrings(t *testing.T) {
	qsortInstead = qsortNever

	l := []string{"zomg", "stuff", "things", "blah", "asdf"}

	if StringsAreSorted(l) {
		t.Error("unsorted string list reports as sorted")
	}

	radix(StringSlice(l))

	if !StringsAreSorted(l) {
		t.Error("after sorting string list reports as not sorted")
		t.Log(l)
	}

	if SearchStrings(l, "zomg") != 4 {
		t.Error("binary search failed for string list")
	}
}
