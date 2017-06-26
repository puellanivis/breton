package wrapper

import (
	"net/url"
	"os"
	"sync"
	"time"
)

type Info struct {
	sync.Mutex

	name string
	uri *url.URL
	sz int64
	t time.Time
}

func (fi *Info) Stat() (os.FileInfo, error) {
	return fi, nil
}

func (fi *Info) Name() string {
	fi.Lock()
	defer fi.Unlock()

	if fi.name == "" && fi.uri != nil {
		fi.name = fi.uri.String()
	}
	  
	return fi.name
}

func (fi *Info) Size() int64 {
	return fi.sz
}

func (fi *Info) Mode() os.FileMode {
	return os.FileMode(0644)
}

func (fi *Info) ModTime() time.Time {
	return fi.t
}

func (fi *Info) IsDir() bool {
	return false
}

func (fi *Info) Sys() interface{} {
	return nil
}

func NewInfo(uri *url.URL, size int, t time.Time) *Info {
	return &Info{
		uri: uri,
		sz: int64(size),
		t: t,
	}
}
