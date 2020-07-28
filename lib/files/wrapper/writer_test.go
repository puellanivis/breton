package wrapper

import (
	"testing"

	"github.com/puellanivis/breton/lib/files"
)

func TestImplementsFilesWriter(t *testing.T) {
	var f files.SyncWriter = new(Writer)

	_ = f
}
