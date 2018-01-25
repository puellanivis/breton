package glog

import (
	"bytes"
	"strconv"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// getBuffer returns a new, ready-to-use buffer.
func getBuffer() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

// putBuffer returns a buffer to the free list.
func putBuffer(b *bytes.Buffer) {
	if b.Len() >= 256 {
		// Let big buffers die a natural death.
		return
	}

	bufPool.Put(b)
}

func writeTwo(buf *bytes.Buffer, i int) {
	if i < 10 {
		buf.WriteByte('0')
	}
	buf.WriteString(strconv.Itoa(i))
}

func ensureNL(buf *bytes.Buffer) {
	b := buf.Bytes()
	if b[len(b)-1] != '\n' {
		buf.WriteByte('\n')
	}
}
