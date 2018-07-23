package buffer

const segmentCapacity = 0x100

type segment struct {
	b [segmentCapacity]byte
}

func (s *segment) Copy() *segment {
	n := new(segment)
	copy(n.b[:], s.b[0:s.Len()+1])
	return n
}

func (s *segment) Len() int {
	return int(s.b[0])
}

func (s *segment) Cap() int {
	return segmentCapacity
}

func (s *segment) Trunc(n int) int {
	l := s.Len()

	if n >= l {
		return l
	}

	s.b[0] = byte(n)
	return n
}

func (s *segment) Bytes() []byte {
	return s.b[1 : s.Len()+1]
}

func (s *segment) Append(b []byte) int {
	rem := s.Cap() - s.Len()

	if len(b) > rem {
		// buffer is too long, truncate the slice
		b = b[:rem]
	}

	n := copy(s.b[s.Len()+1:], b)

	s.b[0] += byte(n)

	return n
}

func (s *segment) Tail(i int) *segment {
	l := s.Len()

	if i >= l {
		return nil
	}

	n := new(segment)
	n.Append(s.b[i+1 : l+1])

	return n
}
