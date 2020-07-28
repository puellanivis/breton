package s3files

import (
	"testing"

	"github.com/puellanivis/breton/lib/files"
)

func TestHandlerFulfillsReadDirFS(t *testing.T) {
	var h files.FS = &handler{}

	if _, ok := h.(files.ReadDirFS); !ok {
		t.Fatal("handler does not implement files.ReadDirFS")
	}

	if _, ok := h.(files.CreateFS); !ok {
		t.Fatal("handler does not implement files.CreateFS")
	}
}
