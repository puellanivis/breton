package socketfiles

import (
	"context"
	"net"
	"net/url"
	"testing"
)

func TestUDPName(t *testing.T) {
	w := &udpWriter{
		ipSocket: ipSocket{
			laddr: &net.UDPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 65535,
			},
			raddr: &net.UDPAddr{
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
		buf: make([]byte, 188),
	}

	uri := w.uri()
	expected := "udp://127.0.0.1:80?buffer_size=1024&localaddr=127.0.0.1&localport=65535&max_bitrate=2048&pkt_size=188&tos=0x80&ttl=100"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}

	w = &udpWriter{
		ipSocket: ipSocket{
			laddr: &net.UDPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 65534,
			},
			raddr: &net.UDPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 443,
			},
		},
	}

	uri = w.uri()
	expected = "udp://127.0.0.1:443?localaddr=127.0.0.1&localport=65534"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}
}

func TestUDPNoSlashSlash(t *testing.T) {
	uri, err := url.Parse("udp:0.0.0.0:0")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL")
	}

	ctx, cancel := context.WithCancel(context.Background())

	_, err = (&udpHandler{}).Open(ctx, uri)
	if err == nil {
		t.Fatal("expected Open(\"udp:0.0.0.0:0\") to error, it did not")
	}
	cancel()

	ctx, cancel = context.WithCancel(context.Background())

	_, err = (&udpHandler{}).Create(ctx, uri)
	if err == nil {
		t.Fatal("expected Open(\"udp:0.0.0.0:0\") to error, it did not")
	}
	cancel()
}

func TestUDPEmptyURL(t *testing.T) {
	uri, err := url.Parse("udp:")
	if err != nil {
		t.Fatal("unexpected error parsing constant URL")
	}

	ctx, cancel := context.WithCancel(context.Background())

	_, err = (&udpHandler{}).Open(ctx, uri)
	if err == nil {
		t.Fatal("expected Open(\"udp:0.0.0.0:0\") to error, it did not")
	}
	cancel()

	ctx, cancel = context.WithCancel(context.Background())

	_, err = (&udpHandler{}).Create(ctx, uri)
	if err == nil {
		t.Fatal("expected Open(\"udp:0.0.0.0:0\") to error, it did not")
	}
	cancel()
}
