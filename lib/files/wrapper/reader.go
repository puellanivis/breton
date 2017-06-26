package wrapper

import (
	"bytes"
	"net/url"
	"os"
	"sync"
	"time"
)

// Reader implements files.Reader with an underlying byte slice.
type Reader struct {
	sync.Mutex

	*Info
	b *bytes.Reader
}

// NewReaderWithInfo returns a new Reader with the given FileInfo.
func NewReaderWithInfo(info os.FileInfo, b []byte) *Reader {
	inf, ok := info.(*Info)
	if !ok {
		inf = &Info{
			name: info.Name(),
			sz:   info.Size(),
			t:    info.ModTime(),
		}
	}

	return &Reader{
		b:    bytes.NewReader(b),
		Info: inf,
	}
}

// NewReader returns a new Reader with a NewInfo with uri, len(b) and time.Time specified.
func NewReader(uri *url.URL, b []byte, t time.Time) *Reader {
	return NewReaderWithInfo(NewInfo(uri, len(b), t), b)
}

// Read performs a thread-safe Read from the underlying Reader.
func (r *Reader) Read(b []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.b.Read(b)
}

// Seek performs a thread-safe Seek to the underlying Reader.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	r.Lock()
	defer r.Unlock()

	return r.b.Seek(offset, whence)
}

// Close recovers resources assigned in the Reader.
func (r *Reader) Close() error {
	r.Lock()
	defer r.Unlock()

	r.b = nil
	r.Info = nil
	return nil
}
