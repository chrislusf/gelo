package extensions

import (
	"bufio"
	"bytes"
	"gelo"
	"os"
)

var (
	Stdio  gelo.Port = &_stdio{bufio.NewReader(os.Stdin)}
	Stderr gelo.Port = &_stderr{}
)

type _stdio struct {
	*bufio.Reader
}

func (s *_stdio) Send(w gelo.Word) {
	var out []byte
	if l, ok := w.(*gelo.List); ok {
		var buf bytes.Buffer
		for ; l.Next != nil; l = l.Next {
			buf.Write(l.Value.Ser().Bytes())
			buf.WriteString(" ")
		}
		buf.Write(l.Value.Ser().Bytes())
		out = buf.Bytes()
	} else {
		out = w.Ser().Bytes()
	}
	os.Stdout.Write(out)
	os.Stdout.WriteString("\n")
}

func (s *_stdio) Recv() gelo.Word {
	line, _ := s.Reader.ReadBytes('\n')
	return gelo.BytesToSym(line[0 : len(line)-1])
}

//Cannot close stdio, should be an error but don't have vm access so it would
//only be mysterious
func (s *_stdio) Close() {}

func (s *_stdio) Closed() bool {
	return false
}

func (s *_stdio) Ser() gelo.Symbol {
	return s.Type()
}

func (s *_stdio) Copy() gelo.Word {
	return s
}

func (s *_stdio) DeepCopy() gelo.Word {
	return s
}

func (s *_stdio) Equals(w gelo.Word) bool {
	_, ok := w.(*_stdio)
	return ok
}

var _stdio_t = gelo.StrToSym("*STDIO*")

func (*_stdio) Type() gelo.Symbol {
	return _stdio_t
}

type _stderr struct{}

func (s *_stderr) Send(w gelo.Word) {
	os.Stderr.Write(w.Ser().Bytes())
	os.Stderr.WriteString("\n")
}

func (s *_stderr) Recv() gelo.Word {
	return gelo.Null
}

//Cannot close stderr, should be an error but don't have vm access so it would
//only be mysterious
func (s *_stderr) Close() {}

func (s *_stderr) Closed() bool {
	return false
}

func (s *_stderr) Ser() gelo.Symbol {
	return s.Type()
}

func (s *_stderr) Copy() gelo.Word {
	return s
}

func (s *_stderr) DeepCopy() gelo.Word {
	return s
}

func (s *_stderr) Equals(w gelo.Word) bool {
	_, ok := w.(*_stderr)
	return ok
}

var _stderr_t = gelo.StrToSym("*STDIO*")

func (*_stderr) Type() gelo.Symbol {
	return _stderr_t
}
