package wrapper

import (
	"bytes"
	"net/url"
	"os"
	"sync"
	"time"
)

type Reader struct {
	sync.Mutex

	*Info
	b *bytes.Reader
}

func NewReaderWithInfo(info os.FileInfo, b []byte) *Reader {
	inf, ok := info.(*Info)
	if !ok {
		inf = &Info{
		name: info.Name(),
		sz: info.Size(),
		t: info.ModTime(),
		}
	}

	return &Reader{
		b: bytes.NewReader(b),
		Info: inf,
	}
}

func NewReader(uri *url.URL, b []byte, t time.Time) *Reader {
	return NewReaderWithInfo(NewInfo(uri, len(b), t), b)
}

func (r *Reader) Read(b []byte) (int, error) {
	r.Lock()
	defer r.Unlock()

	return r.b.Read(b)
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	r.Lock()
	defer r.Unlock()

	return r.b.Seek(offset, whence)
}

func (r *Reader) Close() error {
	r.Lock()
	defer r.Unlock()

	r.b = nil
	r.Info = nil
	return nil
}
