package s3files

import (
	"bytes"
	"context"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (h *handler) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	bucket, key, err := getBucketKey(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "create",
			Path: uri.String(),
			Err:  err,
		}
	}

	path := uri.String()

	return wrapper.NewWriter(ctx, uri, func(b []byte) error {
		cl, err := h.getClient(ctx, bucket)
		if err != nil {
			return &os.PathError{
				Op:   "write",
				Path: path,
				Err:  err,
			}
		}

		req := &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader(b),
		}

		_, err = cl.PutObjectWithContext(ctx, req)
		if err != nil {
			return &os.PathError{
				Op:   "put_object",
				Path: path,
				Err:  err,
			}
		}

		return nil
	}), nil
}
