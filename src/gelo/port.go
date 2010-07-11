package gelo

type Chan struct {
	c chan Word
}

func NewChan() Port {
	return &Chan{make(chan Word)}
}

func (c *Chan) Send(w Word) {
	c.c <- w.DeepCopy()
}

func (c *Chan) Recv() Word {
	w := <-c.c
	if w == nil { //XXX
		if _, ok := w.(*List); ok {
			return EmptyList
		}
		//otherwise the channel has been closed
		return Null
	}
	return w.DeepCopy()
}

func (c *Chan) Close() {
	close(c.c)
}

func (c *Chan) Closed() bool {
	return closed(c.c)
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
	return oc.c == c.c
}

func (c *Chan) Type() Symbol {
	return interns("*CHAN*")
}
