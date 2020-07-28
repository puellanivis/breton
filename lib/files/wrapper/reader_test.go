package wrapper

import (
	"testing"

	"github.com/puellanivis/breton/lib/files"
)

func TestImplementsFilesReader(t *testing.T) {
	var f files.SeekReader = new(Reader)

	_ = f
}
