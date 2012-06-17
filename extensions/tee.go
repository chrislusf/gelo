package extensions

import (
	"bytes"
	"code.google.com/p/gelo"
)

//BUG(jmf): would not behave well if ith Port is Closed()

type tee []gelo.Port

func Tee(ports ...gelo.Port) gelo.Port {
	t := tee(ports)
	return &t
}

func (t *tee) Ser() gelo.Symbol {
	//Tee{p0 p1 ... pN}
	var buf bytes.Buffer
	buf.WriteString("Tee{")
	for _, p := range *t {
		buf.Write(p.Ser().Bytes())
	}
	buf.WriteByte('}')
	return gelo.BytesToSym(buf.Bytes())
}

func (t *tee) Copy() gelo.Word {
	out := tee(make([]gelo.Port, len(*t)))
	for i, p := range *t {
		out[i] = p
	}
	return &out
}

func (t *tee) DeepCopy() gelo.Word {
	out := tee(make([]gelo.Port, len(*t)))
	for i, p := range *t {
		out[i] = p.DeepCopy().(gelo.Port)
	}
	return &out
}

func (t *tee) Equals(o gelo.Word) bool {
	ot, ok := o.(*tee)
	if !ok {
		return false
	}
	for i, p := range *t {
		if !(*ot)[i].Equals(p) {
			return false
		}
	}
	return true
}

func (t *tee) Type() gelo.Symbol {
	return gelo.StrToSym("*TEE-PORT*")
}

func (t *tee) Send(w gelo.Word) {
	for _, p := range *t {
		p.Send(w)
	}
}

// If closed, return Null, if one port return t[0].Recv(), otherwise
//return a list L where L_i = (ith port).Recv()
func (t *tee) Recv() gelo.Word {
	switch len(*t) {
	case 0:
		return gelo.Null
	case 1:
		return (*t)[0].Recv()
	}
	lb := ListBuilder()
	for _, p := range *t {
		lb.Push(p.Recv())
	}
	return lb.List()
}

func (t *tee) Close() {
	for i, p := range *t {
		p.Close()
		(*t)[i] = nil
	}
	*t = (*t)[:0]
}

func (t *tee) Closed() bool {
	return len(*t) == 0
}
