// Package s3files implements the "s3:" URL scheme.
package s3files

import (
	"context"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type region struct {
	region string
	enc    bool

	sess *session.Session
	cl   *s3.S3
}

type handler struct {
	mu sync.Mutex

	defRegion string
	rmap      map[string]*region
}

const defaultRegion = "us-east-1"

func init() {
	h := &handler{
		defRegion: defaultRegion,
		rmap:      make(map[string]*region),
	}

	files.RegisterScheme(h, "s3")
}

func newRegion(r string) (*region, error) {
	conf := &aws.Config{
		Region: aws.String(r),
	}

	sess, err := session.NewSession(conf)
	if err != nil {
		return nil, err
	}

	return &region{
		region: r,
		sess:   sess,
		cl:     s3.New(sess, conf),
	}, nil
}

func (h *handler) getClient(ctx context.Context, bucket string) (*s3.S3, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	region := h.defRegion

	if i := strings.LastIndexByte(bucket, '.'); i >= 0 {
		bucket, region = bucket[:i], bucket[i+1:]
	}

	var err error

	r := h.rmap[region]
	if r == nil {
		r, err = newRegion(region)
		if err != nil {
			return nil, err
		}
		h.rmap[region] = r
	}

	region, err = s3manager.GetBucketRegion(ctx, r.sess, bucket, region)
	if err != nil {
		return nil, err
	}

	r = h.rmap[region]
	if r == nil {
		r, err = newRegion(region)
		if err != nil {
			return nil, err
		}
		h.rmap[region] = r
	}

	return r.cl, nil
}

func (h *handler) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	bucket := uri.Host
	key := uri.Path

	if bucket == "" || key == "" {
		return nil, &os.PathError{"list", uri.String(), os.ErrInvalid}
	}

	cl, err := h.getClient(ctx, bucket)
	if err != nil {
		return nil, &os.PathError{"list", uri.String(), err}
	}

	req := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	}

	res, err := cl.ListObjectsWithContext(ctx, req)
	if err != nil {
		return nil, &os.PathError{"list", uri.String(), err}
	}

	var fi []os.FileInfo
	for _, o := range res.Contents {
		var name string
		if o.Key != nil {
			name = *o.Key
		}

		var sz int64
		if o.Size != nil {
			sz = *o.Size
		}

		var lm time.Time
		if o.LastModified != nil {
			lm = *o.LastModified
		}

		uri := &url.URL{
			Scheme: uri.Scheme,
			Host:   bucket,
			Path:   name,
		}

		fi = append(fi, wrapper.NewInfo(uri, int(sz), lm))
	}

	return fi, nil
}
