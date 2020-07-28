package files

import (
	"context"
	"net/url"
	"reflect"
	"runtime"
	"testing"
)

func TestInvalidURLAsSimplePath(t *testing.T) {
	path := parsePath(context.Background(), ":/foo")
	expect := &url.URL{
		Path: ":/foo",
	}

	if !reflect.DeepEqual(path, expect) {
		t.Errorf("parsePath returned %#v, expected %#v", path, expect)
	}
}

func TestPathWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}

	ctx := context.Background()

	root := parsePath(ctx, "C:\\")
	expect := &url.URL{
		Path: "C:\\",
	}

	if !reflect.DeepEqual(root, expect) {
		t.Fatalf("parsePath returned %#v, expected: %#v", root, expect)
	}

	ctx = WithRootURL(ctx, root)

	filename := parsePath(ctx, "filename")

	expect = &url.URL{
		Path: "C:\\filename",
	}

	if !reflect.DeepEqual(filename, expect) {
		t.Errorf("resolveFilename returned %#v, expected: %#v", filename, expect)
	}
}

func TestPathPOSIX(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}

	ctx := context.Background()

	root := parsePath(ctx, "/tmp")
	expect := &url.URL{
		Path: "/tmp",
	}

	if !reflect.DeepEqual(root, expect) {
		t.Fatalf("parsePath returned %#v, expected: %#v", root, expect)
	}

	ctx = WithRootURL(ctx, root)

	filename := parsePath(ctx, "filename")

	expect = &url.URL{
		Path: "/tmp/filename",
	}

	if !reflect.DeepEqual(filename, expect) {
		t.Errorf("resolveFilename returned %#v, expected: %#v", filename, expect)
	}
}

func TestPathURL(t *testing.T) {
	ctx := context.Background()

	path := "scheme://username:password@hostname:12345/path/?query#fragment"

	root := parsePath(ctx, path)
	expect := path

	if got := root.String(); got != expect {
		t.Fatalf("parsePath returned %q, expected: %q", root, expect)
	}

	ctx = WithRootURL(ctx, root)

	filename := parsePath(ctx, "filename?newquery#newfragment")
	expect = "scheme://username:password@hostname:12345/path/filename?newquery#newfragment"

	if got := filename.String(); got != expect {
		t.Errorf("resolveFilename returned %q, expected: %q", filename, expect)
	}
}
