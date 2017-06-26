package files

import (
	"io"
	"os"
)

type File interface {
	io.Closer
	Name() string
	Stat() (os.FileInfo, error)
}

type Reader interface {
	File
	Seek(int64, int) (int64, error)
	io.Reader
}

type Writer interface {
	File
	io.Writer
	Sync() error
}
