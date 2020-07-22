package files

import (
	"context"
	"net/url"
	"os"
)

// FileStore defines an interface which implements a system of accessing files for reading (Open) writing (Write) and directly listing (List)
type FileStore interface {
	Open(ctx context.Context, uri *url.URL) (Reader, error)
	Create(ctx context.Context, uri *url.URL) (Writer, error)
	List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error)
}
