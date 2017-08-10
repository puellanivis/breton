// Package wrapper provides a files.Files implementation based on a bytes.Buffer backing store, and WriteFn callbacks.
package wrapper

import (
	"net/url"
	"os"
	"time"
)

// Info provides an implementation of os.FileInfo with arbitrary information suitable for a url.
type Info struct {
	name string
	uri  *url.URL
	sz   int64
	mode os.FileMode
	t    time.Time
}

// NewInfo returns a new Info set with the url, size and time specified.
func NewInfo(uri *url.URL, size int, t time.Time) *Info {
	return &Info{
		uri:  uri,
		sz:   int64(size),
		mode: os.FileMode(0644),
		t:    t,
	}
}

// Stat returns the Info as an os.FileInfo, required for implementation of files.File
func (fi *Info) Stat() (os.FileInfo, error) {
	return fi, nil
}

// Name returns the filename of the Info, if name == "" and there is a url,
// then it renders the url, and returns that as the name.
func (fi *Info) Name() string {
	if fi.name == "" && fi.uri != nil {
		return fi.uri.String()
	}

	return fi.name
}

// Size returns the size declared in the Info.
func (fi *Info) Size() int64 {
	return fi.sz
}

// Mode returns the last value set via Chmod(), this defaults to os.FileMode(0644)
func (fi *Info) Mode() os.FileMode {
	return fi.mode
}

// Chmod sets the os.FileMode to be returned from Mode().
func (fi *Info) Chmod(mode os.FileMode) error {
	fi.mode = mode
	return nil
}

// ModTime returns the modification time declared in the Info.
func (fi *Info) ModTime() time.Time {
	return fi.t
}

// IsDir returns false. No Info object should be a directory.
func (fi *Info) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys returns nil, it could potentially later hold the actual underyling buffer...
func (fi *Info) Sys() interface{} {
	return fi
}
