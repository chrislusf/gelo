package extensions

import (
	"gelo"
	"bytes"
)

type couple []gelo.Port

//create a port that reads from in and writes to out
func Couple(in, out gelo.Port) gelo.Port {
	t := couple([]gelo.Port{in, out})
	return &t
}

func (c *couple) Ser() gelo.Symbol {
	//Couple{p_in p_out} or Couple{} if closed
	var buf bytes.Buffer
	buf.WriteString("Couple{")
	for _, p := range *c {
		buf.Write(p.Ser().Bytes())
	}
	buf.WriteByte('}')
	return gelo.BytesToSym(buf.Bytes())
}

func (c *couple) Copy() gelo.Word {
	out := couple(make([]gelo.Port, len(*c)))
	for i, p := range *c {
		out[i] = p
	}
	return &out
}

func (c *couple) DeepCopy() gelo.Word {
	out := couple(make([]gelo.Port, len(*c)))
	for i, p := range *c {
		out[i] = p.DeepCopy().(gelo.Port)
	}
	return &out
}

func (c *couple) Equals(o gelo.Word) bool {
	oc, ok := o.(*couple)
	if !ok {
		return false
	}
	for i, p := range *c {
		if !(*oc)[i].Equals(p) {
			return false
		}
	}
	return true
}

func (c *couple) Type() gelo.Symbol {
	return gelo.StrToSym("*COUPLED-PORT*")
}

func (c *couple) Send(w gelo.Word) {
	if c.Closed() {
		return
	}
	(*c)[1].Send(w)
}

func (c *couple) Recv() gelo.Word {
	if c.Closed() {
		return gelo.Null
	}
	return (*c)[0].Recv()
}

func (c *couple) Close() {
	if c.Closed() {
		return
	}
	(*c)[0].Close()
	(*c)[1].Close()
	*c = (*c)[:0]
}


func (c *couple) Closed() bool {
	return len(*c) == 0
}
