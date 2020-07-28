package cachefiles

import (
	"testing"

	"github.com/puellanivis/breton/lib/files"
)

func TestHandlerFulfillsReadDirFS(t *testing.T) {
	var h files.FS = &FS{}

	if _, ok := h.(files.ReadDirFS); !ok {
		t.Fatal("handler does not implement files.ReadDirFS")
	}
}

func TestHandlerFulfillsCreateFS(t *testing.T) {
	var h files.FS = &FS{}

	if _, ok := h.(files.CreateFS); !ok {
		t.Fatal("handler does not implement files.CreateFS")
	}
}
