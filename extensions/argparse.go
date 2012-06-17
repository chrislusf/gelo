package extensions

import (
	"gelo"
	"bytes"
)

type __token byte

const (
	_eoi __token = iota
	_eot
	_alt
	_lit
	_star
	_plus
	_maybe
	_oseq
	_cseq
	_char
)

type __arg_parserparser struct {
	s, l map[string]bool
	spec []byte
	ch   __token
	pos  int
	buf  *bytes.Buffer
}

type _kv struct {
	key  string
	val  gelo.Word
	next *_kv
}

type _match struct {
	head, tail *_kv
	cont       *gelo.List
}

type _arg_parser interface {
	match(*gelo.List, *__arg_parserparser) (bool, *_match)
}

type _lit_parser string
type _var_parser string
type _maybe_parser struct {
	p _arg_parser
}
type _plus_parser struct {
	p _arg_parser
}
type _star_parser struct {
	p _arg_parser
}
type _seq_parser struct {
	p    _arg_parser
	next *_seq_parser
}
type _alt_parser struct {
	p    _arg_parser
	next *_alt_parser
}

func (seq *_seq_parser) match(l *gelo.List, pp *__arg_parserparser) (bool, *_match) {
	if l == nil {
		return false, nil
	}
	var head, tail *_kv
	for p := seq; p != nil; p = p.next {
		ok, m := p.p.match(l, pp)
		if !ok {
			return false, nil
		}
		//to allow for parsers that can succeed without a match (*?)
		if m != nil {
			l = m.cont //note that if m == nil, l stays put so "a? b?" will pass
			if head != nil {
				tail.next = m.head
				tail = m.tail
			} else {
				head = m.head
				tail = m.tail
			}
		}
	}
	return true, &_match{head, tail, l}
}

func (alt *_alt_parser) match(l *gelo.List, pp *__arg_parserparser) (bool, *_match) {
	if l == nil {
		return false, nil
	}
	save := l
	for p := alt; p != nil; p = p.next {
		ok, m := p.p.match(l, pp)
		if ok && m != nil {
			//a* and a? never fail but don't necessarily return a match
			return true, m
		}
		l = save //reset position pointer
	}
	return false, nil
}

func (lit _lit_parser) match(l *gelo.List, pp *__arg_parserparser) (bool, *_match) {
	if l == nil {
		return false, nil
	}
	sym := string(lit)
	if bytes.Equal([]byte(sym), l.Value.Ser().Bytes()) {
		kv := &_kv{sym, gelo.Null, nil}
		return true, &_match{kv, kv, l.Next}
	}
	return false, nil
}

func (v _var_parser) match(l *gelo.List, pp *__arg_parserparser) (bool, *_match) {
	if l == nil {
		return false, nil
	}
	kv := &_kv{string(v), l.Value, nil}
	return true, &_match{kv, kv, l.Next}
}

func (m *_maybe_parser) match(l *gelo.List, pp *__arg_parserparser) (bool, *_match) {
	if l == nil {
		return true, nil
	}
	_, r := m.p.match(l, pp)
	return true, r
}

func (s *_star_parser) match(l *gelo.List, pp *__arg_parserparser) (bool, *_match) {
	if l == nil {
		return true, nil
	}
	//consume rest of args
	var head, tail *_kv
	for l != nil {
		ok, m := s.p.match(l, pp)
		//could have something like ['a b]+
		if !ok {
			break
		}
		if m != nil {
			if head != nil {
				tail.next = m.head
				tail = m.tail
			} else {
				head = m.head
				tail = m.tail
			}
			l = m.cont
		}
	}
	if head == nil {
		return true, nil
	}
	//ensure that no matter how convoluted the subparsers are any variables
	//that are not symbols are marked as lists even
	for mark := head; mark != nil; mark = mark.next {
		if !pp.s[mark.key] {
			pp.l[mark.key] = true
		}
	}
	return true, &_match{head, tail, l}
}

func (p *_plus_parser) match(l *gelo.List, pp *__arg_parserparser) (bool, *_match) {
	if l == nil {
		return false, nil
	}
	ok, m := p.p.match(l, pp)
	if !ok {
		return false, nil
	}
	_, star := (*_star_parser)(p).match(m.cont, pp)
	if star != nil {
		m.tail.next = star.head
		m.tail = star.tail
		m.cont = star.cont
	}
	for mark := m.head; mark != nil; mark = mark.next {
		if !pp.s[mark.key] {
			pp.l[mark.key] = true
		}
	}
	return true, m
}

func (pp *__arg_parserparser) _next() {
	pp.pos++
	if pp.ch == _eoi || pp.pos >= len(pp.spec) {
		pp.ch = _eoi
		return
	}
	ch := pp.spec[pp.pos]
	if int(ch) < 32 {
		gelo.SyntaxError("Illegal character in argument parser specification")
	}
	switch ch {
	case '\\':
		pp._next() //This is severly broken, what was I thinking?
		if pp.ch == _eoi {
			gelo.SyntaxError("Attempted to escape end of input")
		}
		if pp.ch != _char {
			pp.ch = _char
		}
	case ' ':
		pp.ch = _eot
	case '\'':
		pp.ch = _lit
	case '*':
		pp.ch = _star
	case '+':
		pp.ch = _plus
	case '?':
		pp.ch = _maybe
	case '|':
		pp.ch = _alt
	case '[':
		pp.ch = _oseq
	case ']':
		pp.ch = _cseq
	default:
		pp.ch = _char
	}
	if pp.ch == _char {
		pp.buf.WriteByte(ch)
	}
}

func (pp *__arg_parserparser) _parse_string() string {
	for ; pp.ch == _char; pp._next() {
	}
	out := make([]byte, pp.buf.Len())
	copy(out, pp.buf.Bytes())
	pp.buf.Reset()
	return string(out)
}

func (pp *__arg_parserparser) _parse1(inalt bool) _arg_parser {
	var p _arg_parser
	for ; pp.ch == _eot; pp._next() {
	} //spin past whitespace
	switch pp.ch { //parse prefix
	default:
		gelo.SyntaxError("Unexpected token", string(pp.spec[pp.pos]))
	case _eoi:
		return nil
	case _lit:
		pp._next()
		str := pp._parse_string()
		if len(str) == 0 {
			gelo.SyntaxError("No literal after '")
		}
		pp.s[str] = true
		p = _lit_parser(str)
	case _char:
		//len >= 1, no need for check like above
		p = _var_parser(pp._parse_string())
	case _oseq:
		p = pp._parse_seq(false)
	}
	switch pp.ch {
	default:
		gelo.SyntaxError("Unexpected token", string(pp.spec[pp.pos]))
	case _star, _plus, _maybe:
		switch pp.ch {
		case _star:
			p = &_star_parser{p}
		case _plus:
			p = &_plus_parser{p}
		case _maybe:
			p = &_maybe_parser{p}
		}
		pp._next()
		switch pp.ch {
		case _char, _oseq, _star, _plus, _maybe, _lit:
			gelo.SyntaxError("Unexpected token after postfix operator")
		}
	case _alt:
		if !inalt {
			pp._next()
			p = pp._parse_alt(p)
		}
	case _eot, _eoi, _cseq: //nothing more to do

	}
	return p
}

func (pp *__arg_parserparser) _parse_alt(first _arg_parser) _arg_parser {
	switch pp.ch { //make sure next is kosher
	case _alt, _cseq, _eot, _eoi, _star, _plus, _maybe:
		gelo.SyntaxError("Illegal prefix:", pp.spec[pp.pos])
	}
	head := &_alt_parser{first, nil}
	tail := head
	for {
		if p := pp._parse1(true); p != nil {
			tail.next = &_alt_parser{p, nil}
			tail = tail.next
		} else { //_eoi
			return head
		}
		switch pp.ch {
		default:
			gelo.SyntaxError("Unexpected token",
				string(pp.spec), string(pp.spec[pp.pos-1]))
		case _alt:
			pp._next()
			switch pp.ch { //make sure next is kosher
			case _alt, _cseq, _eot, _star, _plus, _maybe:
				gelo.SyntaxError("Illegal prefix:", pp.spec[pp.pos-1])
			case _eoi:
				gelo.SyntaxError("Input ends after |")
			}
		case _cseq, _eot:
			return head
		}
		pp._next()
	}
	panic("parse alt in impossible state") //Issue 65
}

func (pp *__arg_parserparser) _parse_seq(top bool) *_seq_parser {
	var p _arg_parser
	var head, tail *_seq_parser
	for {
		pp._next()
		if p = pp._parse1(false); p != nil {
			if head != nil {
				tail.next = &_seq_parser{p, nil}
				tail = tail.next
			} else {
				head = &_seq_parser{p, nil}
				tail = head
			}
		}
		switch pp.ch {
		case _cseq:
			if top {
				gelo.SyntaxError("] before [ in argument specification")
			}
			if head == nil {
				gelo.SyntaxError("Empty argument specification")
			}
			pp._next()  //step over ]
			return head //postfix handled in _parse1
		case _eoi:
			if !top {
				gelo.SyntaxError("[ without ] in argument specification")
			}
			if head == nil {
				gelo.SyntaxError("[] empty sequence in argument specification")
			}
			return head
		}
	}
	panic("argument parser parser in impossible state") //Issue 65
}

func _empty_parser(args *gelo.List) (map[string]gelo.Word, bool) {
	if args != nil {
		return nil, false
	}
	return nil, true
}

func MakeArgParser(spec string) func(*gelo.List) (map[string]gelo.Word, bool) {
	if len(spec) == 0 || gelo.SlurpWS([]byte(spec), 0) == len(spec) {
		return _empty_parser
	}
	pp := &__arg_parserparser{
		make(map[string]bool), nil, []byte(spec), _char, -1, new(bytes.Buffer)}
	p := pp._parse_seq(true)
	pp.buf = nil //allow these to be garbage collected
	pp.spec = nil
	//TODO pass map[string]bool around instead of pp
	return func(args *gelo.List) (map[string]gelo.Word, bool) {
		//don't want old matches skewing results so we create an l map each time
		pp.l = make(map[string]bool)
		//attempt to match
		ok, m := p.match(args, pp)
		if !ok || m.cont != nil {
			return nil, false
		}
		defer func() { pp.l = nil }()
		out := make(map[string]gelo.Word)
		repeat := make(map[string]*gelo.List)
		//process matches
		for kv := m.head; kv != nil; kv = kv.next {
			key, val := kv.key, kv.val
			if pp.s[key] { //symbol
				out[key] = gelo.Null
			} else if pp.l[key] { //list
				if tail, ok := repeat[key]; ok {
					tail.Next = &gelo.List{val, nil}
					repeat[key] = tail.Next
				} else { //not seen yet
					l := &gelo.List{val, nil}
					out[key] = l
					repeat[key] = l
				}
			} else { //single var
				if v, ok := out[key]; ok { //promote to list
					pp.l[key] = true
					l := gelo.NewList(v, val)
					out[key] = l
					repeat[key] = l.Next
				} else { //just assign
					out[key] = val
				}
			}
		}
		return out, true
	}
}

func MakeOrElseArgParser(spec string) func(*gelo.VM, *gelo.List) map[string]gelo.Word {
	parser := MakeArgParser(spec)
	return func(vm *gelo.VM, args *gelo.List) map[string]gelo.Word {
		Args, ok := parser(args)
		if !ok {
			if len(spec) != 0 {
				gelo.ArgumentError(vm, "argparser", spec, args)
			} else {
				gelo.ArgumentError(vm, "argparser", "no arguments", args)
			}
		}
		return Args
	}
}
