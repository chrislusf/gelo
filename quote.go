package gelo

import "bytes"

var Noop = &protected_quote{&quote{false, nil, []byte("")}}

func NewQuoteFrom(w Word) Quote {
	if q, ok := w.(Quote); ok {
		return q
	}
	return &protected_quote{&quote{false, nil, dup(w.Ser().Bytes())}}
}

func NewQuoteFromGo(t []byte) Quote {
	return &protected_quote{&quote{false, nil, dup(t)}}
}

//This is a black magic function
func build_quote_from_list(args *List) *quote {
	if args == nil {
		return Noop.unprotect()
	}
	src := newBuf(0)
	var chead, ctail *sNode
	for ; args != nil; args = args.Next {
		src.Write(args.Value.Ser().Bytes())
		src.WriteString(" ")
		if chead != nil {
			//this is safe because when the vm rewrites it will
			//just blindly fill literals, so it doesn't matter what
			//the actual type is
			ctail.next = &sNode{synLiteral, args.Value, nil}
			ctail = ctail.next
		} else {
			chead = &sNode{synLiteral, args.Value, nil}
			ctail = chead
		}
	}
	return &quote{false, &command{chead, nil}, src.Bytes()}
}

//we ONLY call this in very specific situations where we KNOW the quote does not
//parse and want the syntax error for error reporting
func force_synerr(vm *VM, q Quote) (ret *ErrSyntax) {
	Q := q.unprotect()
	defer func() {
		x := recover()
		//we EXPECT there to be an error
		if x == nil {
			systemError(nil, "Assumed\n", q, "had syntax error falsely")
		}
		ret = x.(*ErrSyntax)
	}()
	parse(newBufFrom(Q.source))
	return
}

func (q *quote) fcode() (code *command, ok bool) {
	if q.literal {
		//no need to keep trying to parse over and over
		code, ok = nil, false
	} else if q.code != nil {
		//already parsed
		code, ok = q.code, true
	} else if len(q.source) == 0 {
		//No-op
		code, ok = nil, true
	} else {
		defer func() {
			if x := recover(); x != nil {
				if _, test := x.(*ErrSyntax); test {
					q.literal = true //know it doesn't parse now
					code = nil
					ok = false
					return
				}
				//other errors bubble
				panic(x)
			}
		}()
		q.code = parse(newBufFrom(q.source))
		code, ok = q.code, true
	}
	return
}

type quote struct {
	literal bool //false until proven otherwise
	code    *command
	source  []byte
}

func (q *quote) unprotect() *quote {
	return q
}

func (q *quote) Ser() Symbol {
	return BytesToSym(dup(q.source))
}

func (q *quote) Equals(w Word) bool {
	oq, ok := w.(Quote)
	if !ok {
		return false
	}
	return bytes.Equal(q.source, oq.unprotect().source)
}

func (q *quote) Copy() Word {
	return q
}

func (q *quote) DeepCopy() Word {
	return q
}

func (*quote) Type() Symbol {
	return interns("*QUOTE*")
}

type protected_quote struct {
	protectee *quote
}

func (q *protected_quote) unprotect() *quote {
	return q.protectee
}

func (q *protected_quote) Ser() Symbol {
	return q.protectee.Ser()
}

func (q *protected_quote) Equals(w Word) bool {
	oq, ok := w.(Quote)
	if !ok {
		return false
	}
	return bytes.Equal(q.unprotect().source, oq.unprotect().source)
}


func (q *protected_quote) Copy() Word {
	return q
}

func (q *protected_quote) DeepCopy() Word {
	return q
}

func (*protected_quote) Type() Symbol {
	return interns("*QUOTE*")
}
