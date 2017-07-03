package sort

import (
	"testing"
)

func TestInt32s(t *testing.T) {
	l := []int32{42, 5, 7, 2, 3}

	if Int32sAreSorted(l) {
		t.Error("unsorted int32 list reports as sorted")
	}

	Int32s(l)

	if !Int32sAreSorted(l) {
		t.Error("after sorting int32 list reports as not sorted")
	}

	if SearchInt32s(l, 42) != 4 {
		t.Error("binary search failed for int32 list")
	}
}

func TestReverseInt32s(t *testing.T) {
	a := []int32{42, 5, 7, 2, 3}
	l := Reverse(a)

	if IsSorted(l) {
		t.Error("unsorted reverse int32 list reports as sorted")
	}

	Sort(l)

	if !IsSorted(l) {
		t.Error("after reverse sorting int32 list reports as not sorted")
	}

	if got := SearchFor(l, int32(42)); got != 0 {
		t.Errorf("binary search failed for reversed int32 list got %d wanted 0", got)
	}
}

func TestInt64s(t *testing.T) {
	l := []int64{42, 5, 7, 2, 3}

	if Int64sAreSorted(l) {
		t.Error("unsorted int64 list reports as sorted")
	}

	Sort(l)

	if !Int64sAreSorted(l) {
		t.Error("after sorting int64 list reports as not sorted")
	}

	if SearchInt64s(l, 42) != 4 {
		t.Error("binary search failed for int64 list")
	}
}

func TestFloat32s(t *testing.T) {
	l := []float32{42, 5, 7, 2, 3}

	if Float32sAreSorted(l) {
		t.Error("unsorted int64 list reports as sorted")
	}

	Float32s(l)

	if !Float32sAreSorted(l) {
		t.Error("after sorting int64 list reports as not sorted")
	}

	if SearchFloat32s(l, 42) != 4 {
		t.Error("binary search failed for int64 list")
	}
}
