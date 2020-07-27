package s3files

import (
	"context"
	"net/url"
	"os"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	bucket, key, err := getBucketKey(uri)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	cl, err := h.getClient(ctx, bucket)
	if err != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: uri.String(),
			Err:  err,
		}
	}

	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	res, err := cl.GetObjectWithContext(ctx, req)
	if err != nil {
		return nil, &os.PathError{
			Op:   "get_object",
			Path: uri.String(),
			Err:  err,
		}
	}

	var sz int64
	if res.ContentLength != nil {
		sz = *res.ContentLength
	}

	lm := time.Now()
	if res.LastModified != nil {
		lm = *res.LastModified
	}

	return wrapper.NewReaderWithInfo(res.Body, wrapper.NewInfo(uri, int(sz), lm)), nil
}
