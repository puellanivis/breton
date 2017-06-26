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

func TestInt64s(t *testing.T) {
	l := []int64{42, 5, 7, 2, 3}

	if Int64sAreSorted(l) {
		t.Error("unsorted int64 list reports as sorted")
	}

	Int64s(l)

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
