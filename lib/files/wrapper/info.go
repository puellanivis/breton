// Package wrapper provides a files.Files implementation based on a bytes.Buffer backing store, and WriteFn callbacks.
package wrapper

import (
	"net/url"
	"os"
	"sync"
	"time"
)

// Info provides an implementation of os.FileInfo with arbitrary information suitable for a url.
type Info struct {
	mu sync.Mutex

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

// Name returns the filename of the Info, if name == "" and there is a url,
// then it renders the url, and returns that as the name.
func (fi *Info) Name() string {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	if fi.name == "" && fi.uri != nil {
		fi.name = fi.uri.String()
	}

	return fi.name
}

// Size returns the size declared in the Info.
func (fi *Info) Size() int64 {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	return fi.sz
}

// SetSize sets a new size in the Info, and also updates the ModTime to time.Now()
func (fi *Info) SetSize(size int) {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.sz = int64(size)
}

// Mode returns the last value set via Chmod(), this defaults to os.FileMode(0644)
func (fi *Info) Mode() os.FileMode {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	return fi.mode
}

// Chmod sets the os.FileMode to be returned from Mode().
func (fi *Info) Chmod(mode os.FileMode) error {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.mode = mode
	return nil
}

// ModTime returns the modification time declared in the Info.
func (fi *Info) ModTime() time.Time {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	return fi.t
}

// SetModTime sets the modification time in the Info to the time.Time given.
func (fi *Info) SetModTime(t time.Time) {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.t = t
}

// IsDir returns false. No Info object should be a directory.
func (fi *Info) IsDir() bool {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	return fi.mode&os.ModeDir != 0
}

// Sys returns the Info object itself, as it is already the underlying data source.
func (fi *Info) Sys() interface{} {
	return fi
}

// Stat returns the Info object itself, this allows for a simple embedding of the Info into a struct.
func (fi *Info) Stat() (os.FileInfo, error) {
	return fi, nil
}
