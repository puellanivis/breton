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

var testHandler handler

func TestDataURLs(t *testing.T) {
	type test struct {
		name                string
		input               string
		expected            []byte
		expectedContentType string
	}

	tests := []test{
		{
			name:     "correctly encoded text with default media type",
			input:    "data:,ohai%2A",
			expected: []byte("ohai*"),
		},
		{
			name:     "correctly encoded text with media type",
			input:    "data:type/subtype;foo=bar,ohai%2A",
			expected: []byte("ohai*"),

			expectedContentType: "type/subtype;foo=bar",
		},
		{
			name:     "correctly encoded base64 text with default media type",
			input:    "data:;base64,b2hhaSo=",
			expected: []byte("ohai*"),
		},
		{
			name:     "correctly encoded base64 binary data with default media type",
			input:    "data:;base64,ohai+/Z=",
			expected: []byte{162, 22, 162, 251, 246},
		},
		{
			name:     "correctly encoded base64 with media type",
			input:    "data:type/subtype;foo=bar;base64,b2hhaSo=",
			expected: []byte("ohai*"),

			expectedContentType: "type/subtype;foo=bar",
		},
		{
			name:     "incorrectly encoded base64 directive is actually media type",
			input:    "data:base64,b2hhaSo=",
			expected: []byte("b2hhaSo="),

			expectedContentType: "base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, err := url.Parse(tt.input)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			f, err := testHandler.Open(ctx, uri)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			defer f.Close()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			if !bytes.Equal(b, tt.expected) {
				t.Errorf("got wrong content for %q got %q, wanted %q", tt.input, b, tt.expected)
			}

			h, ok := f.(headerer)
			if !ok {
				t.Fatalf("returned files.Reader does not implement interface{ Header() http.Header }")
			}

			header := h.Header()
			if err != nil {
				t.Fatal("unexpected error:", err)
			}

			expectedContentType := "text/plain;charset=US-ASCII"
			if tt.expectedContentType != "" {
				expectedContentType = tt.expectedContentType
			}

			if got := header.Get("Content-Type"); got != expectedContentType {
				t.Errorf("Content-Type header was %q, expected %q", got, expectedContentType)
			}
		})
	}
}

func TestDataURLNoComma(t *testing.T) {
	uri, err := url.Parse("data:ohai%2A")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := testHandler.Open(ctx, uri)
	if err == nil {
		f.Close()
		t.Fatal("expected error but got none")
	}
}

func TestDataURLWithHost(t *testing.T) {
	uri, err := url.Parse("data://host/,ohai%2A")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := testHandler.Open(ctx, uri)
	if err == nil {
		f.Close()
		t.Fatal("expected error but got none")
	}
}

func TestDataURLWithUser(t *testing.T) {
	uri, err := url.Parse("data://user@/,ohai%2A")
	if err != nil {
		t.Fatal("unexpected error:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	f, err := testHandler.Open(ctx, uri)
	if err == nil {
		f.Close()
		t.Fatal("expected error but got none")
	}
}
