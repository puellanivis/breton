package files

import (
	"io"
	"os"
)

// File defines an interface that abstracts away the concept of files to allow for multiple types of Scheme implementations, and not just local filesystem files.
type File interface {
	io.Closer
	Name() string
	Stat() (os.FileInfo, error)
}

// Reader defines a files.File that is available as an io.ReadSeeker
type Reader interface {
	File
	io.ReadSeeker
}

// Writer defines a files.File that is available as an io.Writer as well as supporting a Sync() function.
type Writer interface {
	File
	io.Writer
	Sync() error
}
