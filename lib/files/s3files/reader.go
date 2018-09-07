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

func (r *reader) Seek(offset int64, whence int) (int64, error) {
	if s, ok := r.ReadCloser.(io.Seeker); ok {
		return s.Seek(offset, whence)
	}

	return 0, &os.PathError{"seek", r.Name(), os.ErrInvalid}
}

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	bucket, key, err := getBucketKey("open", uri)
	if err != nil {
		return nil, err
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

	t := time.Now()
	if res.LastModified != nil {
		t = *res.LastModified
	}

	return &reader{
		ReadCloser: res.Body,
		Info:       wrapper.NewInfo(uri, int(l), t),
	}, nil
}
