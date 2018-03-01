package wrapper

import (
	"testing"

	"github.com/puellanivis/breton/lib/files"
)

func TestImplementsFilesReader(t *testing.T) {
	var f files.Reader = new(Reader)

	_ = f
}
