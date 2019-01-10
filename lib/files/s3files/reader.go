package s3files

import (
	"context"
	"net/url"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (h *handler) Open(ctx context.Context, uri *url.URL) (files.Reader, error) {
	bucket, key, err := getBucketKey("open", uri)
	if err != nil {
		return nil, err
	}

	cl, err := h.getClient(ctx, bucket)
	if err != nil {
		return nil, files.PathError("open", uri.String(), err)
	}

	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	res, err := cl.GetObjectWithContext(ctx, req)
	if err != nil {
		return nil, files.PathError("read", uri.String(), err)
	}

	var l int64
	if res.ContentLength != nil {
		l = *res.ContentLength
	}

	lm := time.Now()
	if res.LastModified != nil {
		lm = *res.LastModified
	}

	return wrapper.NewReaderWithInfo(res.Body, wrapper.NewInfo(uri, int(l), lm)), nil
}
