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
	bucket := uri.Host
	key := uri.Path

	if bucket == "" || key == "" {
		return nil, &os.PathError{"create", uri.String(), os.ErrInvalid}
	}

	region := h.defRegion
	if i := strings.LastIndexByte(bucket, '.'); i >= 0 {
		bucket, region = bucket[:i], bucket[i+1:]
	}

	// The s3files.Writer does not actually perform the request until wrapper.Sync is called,
	// So there is no need for complex synchronization like the s3files.Reader needs.
	w := wrapper.NewWriter(ctx, uri, func(b []byte) error {
		cl, err := h.getClient(ctx, bucket, region)
		if err != nil {
			return &os.PathError{"sync", uri.String(), err}
		}

		req := &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader(b),
		}

		_, err = cl.PutObjectWithContext(ctx, req)
		if err != nil {
			return &os.PathError{"sync", uri.String(), err}
		}

		return nil
	})

	return w, nil
}
