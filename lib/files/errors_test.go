package files

import (
	"errors"
	"os"
	"testing"
)

func TestInvalidURLErrors(t *testing.T) {
	err := &invalidURLError{}

	if !errors.Is(err, os.ErrInvalid) {
		t.Error("errors.Is(invalidURLError, os.ErrInvalid) was false, wanted true")
	}
	if !errors.Is(err, ErrURLInvalid) {
		t.Error("errors.Is(invalidURLError, ErrURLInvalid) was false, wanted true")
	}
	if !errors.Is(err, err) {
		t.Errorf("errors.Is(invalidURLError, invalidURLError) was false, wanted true")
	}

	err2 := errors.Unwrap(err)
	if err2 != ErrURLInvalid {
		t.Errorf("errors.Unwrap(invalidURLError) was %#v, expected %#v", err2, ErrURLInvalid)
	}

	err3 := errors.Unwrap(err2)
	if err3 != os.ErrInvalid {
		t.Errorf("errors.Unwrap(ErrURLInvalid) was %#v, expected %#v", err3, os.ErrInvalid)
	}
}
