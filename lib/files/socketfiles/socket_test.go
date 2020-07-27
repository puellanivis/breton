package socketfiles

import (
	"testing"

	"github.com/puellanivis/breton/lib/files"
)

func TestHandlersFulfillsCreateFS(t *testing.T) {
	var h files.FS = tcpHandler{}

	if _, ok := h.(files.CreateFS); !ok {
		t.Fatal("tcp handler does not implement files.CreateFS")
	}

	h = udpHandler{}
	if _, ok := h.(files.CreateFS); !ok {
		t.Fatal("udp handler does not implement files.CreateFS")
	}

	h = unixHandler{}
	if _, ok := h.(files.CreateFS); !ok {
		t.Fatal("unix handler does not implement files.CreateFS")
	}
}
