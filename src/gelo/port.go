package gelo

import (
	"os"
	"bufio"
)

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

func (c *Chan) Type() Symbol {
	return interns("*CHAN*")
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

type _stdio struct {
	*bufio.Reader
}

func (s *_stdio) Send(w Word) {
	os.Stdout.Write(w.Ser().Bytes())
	os.Stdout.WriteString("\n")
}

func (s *_stdio) Recv() Word {
	line, _ := s.Reader.ReadBytes('\n')
	return BytesToSym(line[0 : len(line)-1])
}

//Cannot close stdio, should be an error but don't have vm access so it would
//only be mysterious
func (s *_stdio) Close() {}

func (s *_stdio) Closed() bool {
	return false
}

func (s *_stdio) Type() Symbol {
	return interns("*STDIO*")
}

func (s *_stdio) Ser() Symbol {
	return s.Type()
}

func (s *_stdio) Copy() Word {
	return s
}

func (s *_stdio) DeepCopy() Word {
	return s
}

func (s *_stdio) Equals(w Word) bool {
	_, ok := w.(*_stdio)
	return ok
}

type _stderr struct {
	read <-chan Word
}

func (s *_stderr) Send(w Word) {
	os.Stderr.Write(w.Ser().Bytes())
}

func (s *_stderr) Recv() Word {
	return <-s.read
}

//Cannot close stderr, should be an error but don't have vm access so it would
//only be mysterious
func (s *_stderr) Close() {}

func (s *_stderr) Closed() bool {
	return false
}

func (s *_stderr) Type() Symbol {
	return interns("*STDIO*")
}

func (s *_stderr) Ser() Symbol {
	return s.Type()
}

func (s *_stderr) Copy() Word {
	return s
}

func (s *_stderr) DeepCopy() Word {
	return s
}

func (s *_stderr) Equals(w Word) bool {
	_, ok := w.(*_stderr)
	return ok
}

var (
	Stdio  = &_stdio{bufio.NewReader(os.Stdin)}
	Stderr = &_stderr{make(<-chan Word)}
)
