package s3files

import (
	"context"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type reader struct {
	io.ReadCloser
	*wrapper.Info
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	bucket := uri.Host
	key := uri.Path

	if bucket == "" || key == "" {
		return nil, &os.PathError{"open", uri.String(), os.ErrInvalid}
	}

	cl, err := h.getClient(ctx, bucket)
	if err != nil {
		return nil, &os.PathError{"open", uri.String(), err}
	}

	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	res, err := cl.GetObjectWithContext(ctx, req)
	if err != nil {
		return nil, &os.PathError{"read", uri.String(), err}
	}

	var l int64
	if res.ContentLength != nil {
		l = *res.ContentLength
	}

	if l < 0 {
		return nil, &os.PathError{"read", uri.String(), os.ErrInvalid}
	}

	b, err := files.ReadFrom(res.Body)
	if err != nil {
		return nil, &os.PathError{"read", uri.String(), err}
	}

	return wrapper.NewReaderFromBytes(b, uri, time.Now()), nil
}
