package wrapper

import (
	"bytes"
	"io"
	"net/url"
	"os"
	"sync"
	"time"
)

// Reader implements files.Reader with an underlying byte slice.
type Reader struct {
	mu sync.Mutex

	fi os.FileInfo
	r  io.Reader
}

// NewReaderWithInfo returns a new Reader with the given FileInfo.
func NewReaderWithInfo(r io.Reader, info os.FileInfo) *Reader {
	return &Reader{
		fi: info,
		r:  r,
	}
}

// NewReaderFromBytes returns a new Reader with a NewInfo with uri, len(b) and time.Time specified.
func NewReaderFromBytes(b []byte, uri *url.URL, t time.Time) *Reader {
	return NewReaderWithInfo(bytes.NewReader(b), NewInfo(uri, len(b), t))
}

// Name implements files.File
func (r *Reader) Name() string {
	return r.fi.Name()
}

// Stat implements files.File
func (r *Reader) Stat() (os.FileInfo, error) {
	return r.fi, nil
}

// Read performs a thread-safe Read from the underlying Reader.
func (r *Reader) Read(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.r.Read(b)
}

// Seek performs a thread-safe Seek to the underlying Reader.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if s, ok := r.r.(io.Seeker); ok {
		return s.Seek(offset, whence)
	}

	return 0, os.ErrInvalid
}

// Close recovers resources assigned in the Reader.
func (r *Reader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var err error

	if r.r == nil {
		return os.ErrClosed
	}

	if c, ok := r.r.(io.Closer); ok {
		err = c.Close()
	}

	r.r = nil

	return err
}
