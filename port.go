package gelo

type Chan struct {
	C      chan Word
	closed bool
}

func NewChan() Port {
	return &Chan{C: make(chan Word)}
}

func (c *Chan) Send(w Word) {
	c.C <- w.DeepCopy()
}

func (c *Chan) Recv() (w Word) {
	if c.closed {
		return Null
	}
	w, c.closed = <-c.C
	if c.closed {
		return Null
	} else if w == nil {
		return EmptyList
	}
	return w.DeepCopy()
}

func (c *Chan) Close() {
	close(c.C)
}

func (c *Chan) Closed() bool {
	return c.closed
}

func (c *Chan) Ser() Symbol {
	return c.Type()
}

func (c *Chan) Copy() Word {
	return c
}

func (c *Chan) DeepCopy() Word {
	return c
}

func (c *Chan) Equals(w Word) bool {
	oc, ok := w.(*Chan)
	if !ok {
		return false
	}
	return oc.C == c.C
}

func (c *Chan) Type() Symbol {
	return interns("*CHAN*")
}
