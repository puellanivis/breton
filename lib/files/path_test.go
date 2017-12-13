package files

import (
	"context"
	"net/url"
	"runtime"
	"testing"
)

func TestPathWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}

	p := makePath("C:\\")
	if !isPath(p) {
		t.Fatalf("makePath returned something not an isPath, got %#v", p)
	}

	if path := getPath(p); path != "C:\\" {
		t.Errorf("getPath(makePath) not inverting, got %s", path)
	}

	ctx := WithRootURL(context.Background(), p)

	filename := makePath("filename")
	if path := resolveFilename(ctx, filename); getPath(path) != "C:\\filename" {
		t.Errorf("resolveFilename with %q and %q gave %#v instead", filename, p, path)
	}
}

func TestPathPOSIX(t *testing.T) {
	if runtime.GOOS == "windows" {
		return
	}

	p := makePath("/asdf")
	if !isPath(p) {
		t.Fatalf("makePath returned something not an isPath, got %#v", p)
	}

	if path := getPath(p); path != "/asdf" {
		t.Errorf("getPath(makePath) not inverting, got %s", path)
	}

	ctx := WithRootURL(context.Background(), p)

	filename := makePath("filename")
	if path := resolveFilename(ctx, filename); getPath(path) != "/asdf/filename" {
		t.Errorf("resolveFilename with %q and %q gave %#v instead", filename, p, path)
	}
}

func TestPathURL(t *testing.T) {
	p, err := url.Parse("scheme://username:password@hostname:port/path/?query#fragment")
	if err != nil {
		t.Fatal(err)
	}

	if isPath(p) {
		t.Fatalf("url.Parse with scheme returned something that is an isPath, got %#v", p)
	}

	ctx := WithRootURL(context.Background(), p)

	filename := makePath("filename")
	if path := resolveFilename(ctx, filename); path.String() != "scheme://username:password@hostname:port/path/filename" {
		t.Errorf("resolveFilename with %q and %q gave %#v instead", filename, p, path)
	}

	p, err = url.Parse("file:///c:/Windows/")
	if err != nil {
		t.Fatal(err)
	}

	if isPath(p) {
		t.Fatalf("url.Parse with scheme returned something that is an isPath, got %#v", p)
	}

	ctx = WithRootURL(context.Background(), p)

	if path := resolveFilename(ctx, filename); path.String() != "file:///c:/Windows/filename" {
		t.Errorf("resolveFilename with %q and %q gave %#v instead", filename, p, path)
	}


}
