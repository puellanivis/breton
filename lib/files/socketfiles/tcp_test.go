package socketfiles

import (
	"context"
	"net"
	"net/url"
	"testing"
	"time"
)

func TestTCPName(t *testing.T) {
	w := &tcpWriter{
		ipSocket: ipSocket{
			laddr: &net.TCPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 65535,
			},
			raddr: &net.TCPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 80,
			},
			bufferSize: 1024,
			ttl:        100,
			tos:        0x80,

			throttler: throttler{
				bitrate: 2048,
			},
		},
	}

	uri := w.uri()
	expected := "tcp://127.0.0.1:80?buffer_size=1024&localaddr=127.0.0.1&localport=65535&max_bitrate=2048&tos=0x80&ttl=100"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}

	w = &tcpWriter{
		ipSocket: ipSocket{
			laddr: &net.TCPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 65534,
			},
			raddr: &net.TCPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 443,
			},
		},
	}

	uri = w.uri()
	expected = "tcp://127.0.0.1:443?localaddr=127.0.0.1&localport=65534"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}
}

func TestTCPNoSlashSlash(t *testing.T) {
	uri, err := url.Parse("tcp:0.0.0.0:0")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	r, err := (&tcpHandler{}).Open(ctx, uri)
	if err == nil {
		r.Close()
		t.Fatal("expected Open(\"tcp:0.0.0.0:0\") to error, it did not")
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)

	w, err := (&tcpHandler{}).Create(ctx, uri)
	if err == nil {
		w.Close()
		t.Fatal("expected Create(\"tcp:0.0.0.0:0\") to error, it did not")
	}
	cancel()
}

func TestTCPEmptyURL(t *testing.T) {
	uri, err := url.Parse("tcp:")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	r, err := (&tcpHandler{}).Open(ctx, uri)
	if err == nil {
		r.Close()
		t.Fatal("expected Open(\"tcp:\") to error, it did not")
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)

	w, err := (&tcpHandler{}).Create(ctx, uri)
	if err == nil {
		w.Close()
		t.Fatal("expected Create(\"tcp:\") to error, it did not")
	}
	cancel()
}

func TestTCPBadLocalAddr(t *testing.T) {
	uri, err := url.Parse("tcp://0.0.0.0:0?localaddr=invalid")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)

	w, err := (&tcpHandler{}).Create(ctx, uri)
	if err == nil {
		w.Close()
		t.Fatal("exepcted Create(\"tcp://0.0.0.0:0?localaddr=invalid\") to error, it did not")
	}
	cancel()
}
