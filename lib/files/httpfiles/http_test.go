package httpfiles

import (
	"testing"

	"github.com/puellanivis/breton/lib/files"
)

func TestHandlerFulfillsCreateFS(t *testing.T) {
	var h files.FS = handler{}

	if _, ok := h.(files.CreateFS); !ok {
		t.Fatal("handler does not implement files.CreateFS")
	}
}
