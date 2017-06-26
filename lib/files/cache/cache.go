package cachefiles

import (
	"net/url"
	"os"
	"sync"
	"time"

	"lib/files"
	"lib/files/wrapper"

	"context"
)

type line struct {
	os.FileInfo

	data []byte
}

type FileStore struct {
	sync.RWMutex

	cache map[string]*line
}

func New() *FileStore {
	return &FileStore{
		cache: make(map[string]*line),
	}
}

var Default = New()

func init() {
	files.RegisterScheme(Default, "cache")
}

func (h *FileStore) expire(filename string) bool {
	h.Lock()
	defer h.Unlock()

	_, ok := h.cache[filename]
	if !ok {
		return false
	}

	delete(h.cache, filename)
	return true
}

func filename(uri *url.URL) string {
	filename := uri.Opaque
	if uri.Opaque == "" {
		filename = uri.String()
		filename = filename[len(uri.Scheme)+1:]
	}

	return filename
}

func (h *FileStore) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	return files.Create(ctx, filename(uri))
}

func (h *FileStore) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	h.Lock()
	defer h.Unlock()

	filename := filename(uri)
	/*if uri.RawQuery != "" {
		filename = fmt.Sprintf("%s?%s", filename, uri.RawQuery)
	}*/

	f, ok := h.cache[filename]

	if !ok {
		if _, ok := ctx.Deadline(); !ok {
			// default 5 minute expire time
			d := 5 * time.Minute
			if t, ok := GetExpire(ctx); ok {
				d = t
			}

			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, d)
			defer cancel()
		}

		raw, err := files.Open(ctx, filename)
		if err != nil {
			return nil, err
		}

		data, err := files.ReadAndClose(raw)
		if err != nil {
			return nil, err
		}

		info, err := raw.Stat()
		if err != nil {
			info = wrapper.NewInfo(uri, len(data), time.Now())
		}

		f = &line{
			data: data,
			FileInfo: info,
		}

		h.cache[filename] = f

		go func() {
			defer h.expire(filename)

			<-ctx.Done()
		}()
	}

	return wrapper.NewReaderWithInfo(f.FileInfo, f.data), nil
}

func (h *FileStore) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return files.List(ctx, filename(uri))
}
