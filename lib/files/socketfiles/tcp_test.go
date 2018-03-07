package socketfiles

import (
	"net"
	"testing"
)

func TestTCPName(t *testing.T) {
	w := &TCPWriter{
		common: common{
			laddr: &net.TCPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 65535,
			},
			raddr: &net.TCPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: 80,
			},
			bufferSize: 1024,
			bitrate:    2048,
			ttl:        100,
			tos:        0x80,
		},
	}

	uri := w.uri()
	expected := "tcp://127.0.0.1:80?bitrate=2048&buffer_size=1024&localaddr=127.0.0.1&localport=65535&tos=0x80&ttl=100"

	if s := uri.String(); s != expected {
		t.Errorf("got a bad URI, was expecting, but got:\n\t%v\n\t%v", expected, s)
	}

	w = &TCPWriter{
		common: common{
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
