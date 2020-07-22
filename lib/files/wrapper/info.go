// Package wrapper provides a files.Files implementation based on a bytes.Buffer backing store, and WriteFn callbacks.
package wrapper

import (
	"net/url"
	"os"
	"sync"
	"syscall"
	"time"
)

// Info provides an implementation of os.FileInfo with arbitrary information suitable for a url.
type Info struct {
	mu sync.RWMutex

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

// SetName sets a new string as the filename.
func (fi *Info) SetName(name string) {
	if fi == nil {
		return
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.name = name
	fi.uri = nil
}

// SetNameFromURL sets a new URL as the filename.
func (fi *Info) SetNameFromURL(uri *url.URL) {
	if fi == nil {
		return
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.uri = uri
	fi.name = ""
}

func (fi *Info) fixName() string {
	if fi.name != "" || fi.uri == nil {
		// Nothing to fix.
		// Likely, someone else already fixed the name while we were waiting on the mutex.
		return fi.name
	}

	fi.name = fi.uri.String()

	if fi.name == "" {
		// If we got an empty string from the url, then we need to remove the url,
		// otherwise we will forever keep trying to fix the name.
		fi.uri = nil
	}

	return fi.name
}

// Name returns the filename of the Info, if name == "" and there is a url,
// then it renders the url, and returns that as the name.
func (fi *Info) Name() string {
	if fi == nil {
		return ""
	}

	fi.mu.RLock()
	name, uri := fi.name, fi.uri
	fi.mu.RUnlock()

	if name == "" && uri != nil {
		fi.mu.Lock()
		defer fi.mu.Unlock()

		return fi.fixName()
	}

	return name
}

// URL returns the URL of the Info, if there is no URL yet,
// then it will set the URL to be `&url.URL{ Path: name }`.
func (fi *Info) URL() *url.URL {
	if fi == nil {
		return &url.URL{}
	}

	fi.mu.RLock()
	uri := fi.uri
	fi.mu.RUnlock()

	if uri != nil {
		u := *uri
		return &u
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	if fi.uri == nil {
		fi.uri = &url.URL{
			Path: fi.name,
		}
	}

	u := *fi.uri
	return &u
}

// Size returns the size declared in the Info.
func (fi *Info) Size() int64 {
	if fi == nil {
		return 0
	}

	fi.mu.RLock()
	defer fi.mu.RUnlock()

	return fi.sz
}

// SetSize sets a new size in the Info.
func (fi *Info) SetSize(size int) {
	if fi == nil {
		return
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.sz = int64(size)
}

// Mode returns the last value set via Chmod(), this defaults to os.FileMode(0644)
func (fi *Info) Mode() (mode os.FileMode) {
	if fi == nil {
		return mode
	}

	fi.mu.RLock()
	defer fi.mu.RUnlock()

	return fi.mode
}

// Chmod sets the os.FileMode to be returned from Mode().
func (fi *Info) Chmod(mode os.FileMode) error {
	if fi == nil {
		return syscall.EINVAL
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.mode = mode
	return nil
}

// ModTime returns the modification time declared in the Info.
func (fi *Info) ModTime() (t time.Time) {
	if fi == nil {
		return t
	}

	fi.mu.RLock()
	defer fi.mu.RUnlock()

	return fi.t
}

// SetModTime sets the modification time in the Info to the time.Time given.
func (fi *Info) SetModTime(t time.Time) {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	fi.t = t
}

// IsDir returns true if a prior Chmod set os.ModeDir.
func (fi *Info) IsDir() bool {
	if fi == nil {
		return false
	}

	fi.mu.RLock()
	defer fi.mu.RUnlock()

	return fi.mode&os.ModeDir != 0
}

// Sys returns the Info object itself, as it is already the underlying data source.
func (fi *Info) Sys() interface{} {
	if fi == nil {
		// return an untyped nil here.
		return nil
	}

	return fi
}

// Stat returns the Info object itself, this allows for a simple embedding of the Info into a struct.
func (fi *Info) Stat() (os.FileInfo, error) {
	// if fi is nil, we intentionally return a typed nil here.
	return fi, nil
}
