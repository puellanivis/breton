package tcpfiles

import (
	"net"
	"testing"
)

func TestName(t *testing.T) {
	w := &Writer{
		laddr: &net.TCPAddr{
			IP: []byte{ 127, 0, 0, 1 },
			Port: 65535,
		},
		raddr: &net.TCPAddr{
			IP: []byte{ 127, 0, 0, 1 },
			Port: 80,
		},
		bufferSize: 1024,
		bitrate: 2048,
		ttl: 100,
		tos: 0x80,
	}

	uri := w.uri()
	expected := "tcp://127.0.0.1:80?buffer_size=1024&localaddr=127.0.0.1&localport=65535&max_bitrate=2048&tos=0x80&ttl=100"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}

	w = &Writer{
		laddr: &net.TCPAddr{
			IP: []byte{ 127, 0, 0, 1 },
			Port: 65534,
		},
		raddr: &net.TCPAddr{
			IP: []byte{ 127, 0, 0, 1 },
			Port: 443,
		},
	}

	uri = w.uri()
	expected = "tcp://127.0.0.1:443?localaddr=127.0.0.1&localport=65534"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}

}
