package datafiles

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type headerer interface {
	Header() http.Header
}

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

	h, ok := f.(headerer)
	if !ok {
		t.Fatalf("returned files.Reader does not implement interface{ Header() (http.Header, error}")
	}

	header := h.Header()
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expectedContentType := "text/plain;charset=US-ASCII"
	if got := header.Get("Content-Type"); got != expectedContentType {
		t.Errorf("unexpected Content-Type header, got %q, wanted %q", got, expectedContentType)
	}
}

func TestDataURLBadBase64(t *testing.T) {
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

	expected := []byte("b2hhaSo=")

	if !bytes.Equal(b, expected) {
		t.Errorf("got wrong content for data:base64,b2hhaSo= got %v, wanted %v", b, expected)
	}
}

func TestDataURLSimpleBase64(t *testing.T) {
	uri, err := url.Parse("data:;base64,b2hhaSo=")
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
	uri, err := url.Parse("data:;base64,ohai+/Z=")
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

func TestDataWithHeader(t *testing.T) {
	uriString := "data:type/subtype;foo=bar;base64,b2hhaSo="
	uri, err := url.Parse(uriString)
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
		t.Errorf("got wrong content for %s got %v, wanted %v", uriString, b, expected)
	}

	h, ok := f.(headerer)
	if !ok {
		t.Fatalf("returned files.Reader does not implement interface{ Header() (http.Header, error}")
	}

	header := h.Header()
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	expectedContentType := "type/subtype;foo=bar"
	if got := header.Get("Content-Type"); got != expectedContentType {
		t.Errorf("unexpected Content-Type header, got %q, wanted %q", got, expectedContentType)
	}
}
