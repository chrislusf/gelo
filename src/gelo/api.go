package gelo

//This type is merely a proxy for the vm to give the API methods a namespace
//distinct from the other vm methods. Pretty, but useless.
type api struct {
	vm *VM
}

func (_ *api) Trace(message ...interface{}) {
	alien_trace(message)
}

func (p *api) Halt(info *List) {
	panic(halt_control_code(info))
}

func (p *api) Recv() Word {
	return p.vm.io.Recv()
}

func (p *api) Send(w Word) {
	p.vm.io.Send(w)
}

//If string, attempt to convert
func (p *api) NumberOrElse(w Word) *Number {
	n, ok := w.(*Number)
	if !ok {
		n, ok = NewNumberFromBytes(w.Ser().Bytes())
		if !ok {
			TypeMismatch(p.vm, "number", w.Type())
		}
	}
	return n
}

func (p *api) QuoteOrElse(w Word) Quote {
	q, ok := w.(Quote)
	if !ok {
		TypeMismatch(p.vm, "quote", w.Type())
	}
	return q
}

func (p *api) ListOrElse(w Word) *List {
	if l, ok := w.(*List); ok {
		return l
	}
	l, ok := UnserializeListFrom(w)
	if !ok {
		TypeMismatch(p.vm, "list", w.Type())
	}
	return l
}

func (p *api) DictOrElse(w Word) *Dict {
	if d, ok := w.(*Dict); ok {
		return d
	}
	d, ok := UnserializeDictFrom(w)
	if !ok {
		TypeMismatch(p.vm, "dict", w.Type())
	}
	return d
}

func (p *api) PortOrElse(w Word) Port {
	port, ok := w.(Port)
	if !ok {
		TypeMismatch(p.vm, "port", w.Type())
	}
	return port
}

func (p *api) ChanOrElse(w Word) Port {
	c, ok := w.(*Chan)
	if !ok {
		TypeMismatch(p.vm, "chan", w.Type())
	}
	return c
}

func (p *api) BoolOrElse(w Word) Bool {
	b, ok := w.(Bool)
	if !ok {
		TypeMismatch(p.vm, "bool", w.Type())
	}
	return b
}

func (p *api) SymbolOrElse(w Word) Symbol {
	s, ok := w.(Symbol)
	if !ok {
		TypeMismatch(p.vm, "symbol", w.Type())
	}
	return s
}

/*
 *      Returns the bytes of a symbol or quote.
 * We do not check that the quote is in fact literal as there are many possible
 * texts that could just happen to parse, this comment being an example of one.
 */
func (p *api) LiteralOrElse(w Word) []byte {
	s, ok := w.(Symbol)
	if ok {
		return s.Bytes()
	}
	q, ok := w.(Quote)
	if !ok {
		TypeMismatch(p.vm, "symbol or quote", w.Type())
	}
	return q.Ser().Bytes()
}

func (p *api) AlienOrElse(w Word) Alien {
	a, ok := w.(Alien)
	if !ok {
		TypeMismatch(p.vm, "alien", w.Type())
	}
	return a
}

func (p *api) InvokeOrElse(args *List) (ret Word) {
	if args == nil {
		SystemError(p.vm, "Invoke* attempted to invoke nothing")
	}
	w, c, args := p.vm.peval(args, uint(args.Len()-1))
	if _, is_defer := w.(defert); is_defer {
		SystemError(p.vm, "Cannot register a defer via Invoke*")
		return
	}
	if c != nil {
		ret = p.vm.eval(c, args)
	} else {
		ret = w
	}
	return
}

func (p *api) Invoke(args *List) (ret Word, err Error) {
	defer func() {
		if x := recover(); x != nil {
			switch t := x.(type) {
			default:
				//note that we are implicitly allowing halt_control_code
				//to bubble
				panic(x)
			case ErrRuntime, ErrSyntax:
				ret, err = nil, x.(Error)
			}
		}
	}()
	ret = p.InvokeOrElse(args)
	return
}

//The TailInvoke* family is only to be called when the result is to be
//returned from the callee.
func (p *api) TailInvoke(args *List) Word {
	if args == nil {
		SystemError(p.vm, "Invoke* attempted to invoke nothing")
	}
	w, c, args := p.vm.peval(args, uint(args.Len()-1))
	if _, is_defer := w.(defert); is_defer {
		SystemError(p.vm, "Cannot register a defer via Invoke*")
	}
	if c != nil {
		if args == nil {
			//we can tail call--either called with no arguments or the return
			//from an alien.
			//TODO, a way to pass arguments back to eval so that we can always
			//tail call.
			return &quote{false, c, nil}
		} else {
			return p.vm.eval(c, args)
		}
	}
	return w
}

func (p *api) InvokeCmd(w Word, args *List) (Word, Error) {
	if q, ok := w.(Quote); ok {
		w = q.unprotect()
	}
	return p.Invoke(&List{w, args})
}

func (p *api) InvokeCmdOrElse(w Word, args *List) Word {
	return p.InvokeOrElse(&List{w, args})
}

func (p *api) TailInvokeCmd(w Word, args *List) Word {
	return p.TailInvoke(&List{w, args})
}

func (p *api) InvokeWordOrReturn(w Word) (ret Word) {
	defer func() {
		if x := recover(); x != nil {
			switch t := x.(type) {
			default:
				panic(x)
			case ErrRuntime, ErrSyntax:
				ret = w
			}
		}
	}()
	switch w.(type) {
	case Quote, Alien, Symbol:
		ret, err := p.Invoke(AsList(w))
		if err == nil {
			return ret
		}
	}
	return w
}

func (p *api) TailInvokeWordOrReturn(w Word) Word {
	if s, ok := w.(Symbol); ok {
		dw, there := p.vm.Ns.Lookup(s)
		if there {
			w = dw
		}
	}
	//manually check if it's code and only unprotect if it is
	if q, ok := w.(Quote); ok {
		uq := q.unprotect()
		if _, ok := uq.fcode(); !ok {
			return q
		}
		w = q
	}
	switch w.(type) {
	case Quote, Alien:
		return p.TailInvoke(AsList(w))
	}
	return w
}

//return a list of lists of symbols and quotes, only evaluates $@[]
func (p *api) PartialEval(q Quote) (*List, bool) {
	cmds, ok := q.unprotect().fcode()
	if !ok { //cannot partially eval what we cannot fully eval
		return nil, false
	}
	if cmds == nil {
		//noop, return singleton containing empty list
		return NewList(EmptyList), true
	}
	var ghead, gtail *List
	for c := cmds; c != nil; c = c.next {
		var head, tail *List
		for s := c.cmd; s != nil; s = s.next {
			var w Word
			switch s.tag {
			case synLiteral, synQuote:
				w = s.val.(Word)
			default:
				//this is ugly but no clean way to extract a rewrite1
				//out of rewrite without increasing the complexity or
				//compromising its performance in the face of a splice.
				//fortunately this is the only place we call it where
				//it wasn't meant to be. Since we handle quote separately,
				//we don't need to worry about it getting unprotected.
				l, _ := p.vm.rewrite(&sNode{s.tag, s.val, nil})
				if s.tag == synSplice {
					if l == nil {
						continue //so we don't write nil to the list
					}
					for ; l.Next != nil; l = l.Next {
					}
					w = l.Next
				} else {
					w = l.Value
				}
			}
			if head != nil {
				tail.Next = &List{w, nil}
				tail = tail.Next
			} else {
				head = &List{w, nil}
				tail = head
			}
		}
		if ghead != nil {
			gtail.Next = &List{head, nil}
			gtail = gtail.Next
		} else {
			ghead = &List{head, nil}
			gtail = ghead
		}
	}
	return ghead, true
}
