package mapreduce

type execChain struct {
	ch      chan struct{}
	ordered bool
}

func newExecChain(ordered bool) *execChain {
	ch := make(chan struct{})
	close(ch)

	return &execChain{
		ch:      ch,
		ordered: ordered,
	}
}

func (c *execChain) next() (ready <-chan struct{}, next chan struct{}) {
	ready = c.ch

	if c.ordered {
		next = make(chan struct{})
		c.ch = next
	}

	return ready, next
}
