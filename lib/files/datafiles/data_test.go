package datafiles

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/url"
	"testing"
	"time"
)

func TestDataURL(t *testing.T) {
	uri, err := url.Parse("data:,ohai%2A")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := (&handler{}).Open(ctx, uri)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := []byte("ohai*")

	if !bytes.Equal(b, expected) {
		t.Errorf("got wrong content for data:,ohai%%2A got %v, wanted %v", b, expected)
	}
}

func TestDataURLSimpleBase64(t *testing.T) {
	uri, err := url.Parse("data:base64,b2hhaSo=")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := (&handler{}).Open(ctx, uri)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := []byte("ohai*")

	if !bytes.Equal(b, expected) {
		t.Errorf("got wrong content for data:base64,b2hhaSo= got %v, wanted %v", b, expected)
	}
}

func TestDataURLComplexBase64(t *testing.T) {
	uri, err := url.Parse("data:base64,ohai+/Z=")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := (&handler{}).Open(ctx, uri)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expected := []byte{162, 22, 162, 251, 246}

	if !bytes.Equal(b, expected) {
		t.Errorf("got wrong content for data:base64,ohai+/Z= got %v, wanted %v", b, expected)
	}
}

func TestDataURLNoComma(t *testing.T) {
	uri, err := url.Parse("data:ohai%2A")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := (&handler{}).Open(ctx, uri)
	if err == nil {
		f.Close()
		t.Fatal("expected error but got none")
	}
}

func TestDataURLWithHost(t *testing.T) {
	uri, err := url.Parse("data://host/,ohai%2A")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := (&handler{}).Open(ctx, uri)
	if err == nil {
		f.Close()
		t.Fatal("expected error but got none")
	}
}

func TestDataURLWithUser(t *testing.T) {
	uri, err := url.Parse("data://user@/,ohai%2A")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := (&handler{}).Open(ctx, uri)
	if err == nil {
		f.Close()
		t.Fatal("expected error but got none")
	}
}
