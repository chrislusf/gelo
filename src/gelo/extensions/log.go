package extensions

import (
	"gelo"
	"log"
	"io"
)

type lr struct {
	*log.Logger
}

//Creates a logger with default prefix "Gelo VM". Use .AsLogger().SetPrefix to
//override
func Logger(out io.Writer, flag int) gelo.Port {
	return &lr{log.New(out, "Gelo VM", flag)}
}

func (g *lr) AsLogger() *log.Logger {
	return g.Logger
}

func (g *lr) Type() gelo.Symbol {
	return gelo.StrToSym("*LOGGER-PORT*")
}

func (g *lr) Ser() gelo.Symbol {
	return g.Type()
}

func (g *lr) Copy() gelo.Word {
	return g
}

func (g *lr) DeepCopy() gelo.Word {
	return g
}

func (g *lr) Equals(o gelo.Word) bool {
	og, ok := o.(*lr)
	return !ok || og == g
}

func (g *lr) Send(w gelo.Word) {
	g.Println(w.Ser().String())
}

func (g *lr) Recv() gelo.Word {
	return gelo.Null
}

func (g *lr) Close() {}

func (g *lr) Closed() bool {
	return false
}
