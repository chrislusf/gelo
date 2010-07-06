package gelo

import "io"

type _ast byte
const (
    synLiteral  _ast = iota
    synSplice
    synIndirect
    synQuote
    synClause
)

type _lexeme byte
const (
    _eol        _lexeme = iota
    _eof
    _l_space    //non-eol whitespace
    _l_indirect //$
    _l_splice   //@
    _l_str      //"
    _lo_clause  //[
    _lc_clause  //]
    _lo_quote   //{
    _lc_quote   //}
    _l_comment  //#
    _l_nil      //everything else
)

type _esc_mode byte
const (
    _reg    _esc_mode = iota
    _str
    _quote
)

var _ctrlchars map[byte]byte = map[byte]byte{
    'n':'\n', 'r':'\r', 'f':'\f', 't':'\t', 'a':'\a', 'b':'\b', 'v':'\v',
}

type sNode struct {
    tag     _ast
    val     interface{}
    next   *sNode
}

type command struct {
    cmd     *sNode
    next    *command
}

type _parser struct {
    record   bool
    escm     _esc_mode
    ch       _lexeme
    cur    []byte
    src      io.Reader
    buf     *buffer
}

func (p *_parser) _adv() {
    if p.ch == _eof {
        return
    }
    _, err := p.src.Read(p.cur)
    if err != nil {
        p.ch = _eof
    }
}

func (p *_parser) _next() {
    if p.ch != _eof && p.record { //don't discard last p.cur
        p.buf.WriteByte(p.cur[0])
    }
    p._adv()
    if p.ch == _eof {
        return
    }
    
    //process escapes
    if p.cur[0] == '\\' {
        p._adv() //get char after \
        if p.ch == _eof {
            SyntaxError("Cannot escape the end of file")
        }
        ch := p.cur[0]

        //we can pick up \* in quotes if and when they get parsed for real
        if p.escm != _quote && ch == '*' {
            //advance cursor to first nonwhitespace
            for {
                p._adv()
                if p.ch == _eof {
                    return
                }
                switch p.cur[0] {
                    case ' ', '\n', '\t', '\f':
                        //slurp
                    default:
                        //we need to process this like any other character now
                        goto normal
                }
            }
        }

        //TODO handle \0xFFAE, etc

        switch p.escm {
            case _reg:
                if esc, ok := _ctrlchars[ch]; ok {
                    //rewrite current ch with escape
                    p.cur[0] = esc
                }
                //otherwise we drop the \ and write the character as a literal
            case _str:
                //always in record mode with strings
                if ch != '"' { //rewrite \" to just "
                    //store and move on
                    p.buf.WriteString("\\")
                }
            case _quote:
                if p.record {
                    //keep all escapes
                    p.buf.WriteString("\\")
                }
        }
        p.ch = _l_nil
        return
    }

    normal:
    switch p.cur[0] {
        case '\n', ';':
            p.ch = _eol
        case ' ', '\t', '\f':
            p.ch = _l_space
        case '#':
            p.ch = _l_comment
        case '$':
            p.ch = _l_indirect
        case '@':
            p.ch = _l_splice
        case '"':
            p.ch = _l_str
        case '[':
            p.ch = _lo_clause
        case ']':
            p.ch = _lc_clause
        case '{':
            p.ch = _lo_quote
        case '}':
            p.ch = _lc_quote
        default: //don't care what it is otherwise
            p.ch = _l_nil
    }
}

func (p *_parser) _read_out() []byte {
    //dischage buffer up to last char read before current ch
    p.record = false
    return p.buf.CopyBytes()
}

func (p *_parser) _parse_word() *sNode {
    var head *sNode
    try_num := false
    join := func(n *sNode) {
        if head != nil {
            head.val = n
        } else {
            head = n
        }
    }
    if p.ch == _l_splice {
        join(&sNode{synSplice, nil, nil})
        p._next()
    } else if p.ch == _l_indirect {
        join(&sNode{synIndirect, nil, nil})
        p._next()
    }
    //handle [] {} "", these conditions only hold if there was a sigil
    switch p.ch {
        case _eol, _eof, _l_space:
            SyntaxError("Sigil precedes nothing")
        case _lo_clause:
            p._next()
            join(p._parse_line(true))
            return head
        case _lo_quote:
            join(p._parse_quote())
            return head
        case _l_str:
            join(p._parse_string())
            return head
    }
    //just a word, slurp till we hit not a word
    p.record = true
    //if first ch is a number or - or + or . we might have a number
    rch := p.cur[0]
    if (48 <= rch && rch <= 57){
        try_num = true
    }
    switch rch {
        case '-', '+', '.':
            try_num = true
    }
    for {
        switch p.ch {
            case _eol, _eof, _l_space, _lo_quote, _lc_quote,
            _lo_clause,  _lc_clause, _l_str:
                r := p._read_out()
                if try_num {
                    n, ok := NewNumberFromGo(r)
                    if ok {
                        n.ser = r //cache serilization
                        join(&sNode{synLiteral, n, nil})
                        return head
                    }
                }
                join(&sNode{synLiteral, intern(r), nil})
                return head
        }
        p._next()
    }
    panic("Failed to parse word on a fundamental level")//Issue 65
}

func (p *_parser) _rquote(record bool) {
    depth := 1
    p.escm = _quote
    //p._next()//step over initial {
    p.record = record
    //since an escape as the first entry in a quote creates special problems
    //some machinery had to be duplicated here to avoid complicating the rest
    p._adv() //eof will be caught in below loop
    if p.ch == _eof {
        SyntaxError("{ without }")
    }
    switch p.cur[0] {
        case '\\':
            p._adv()
            if p.ch == _eof {
                SyntaxError("Cannot escape the end of file")
            }
            p.buf.WriteString("\\")
        case '{':
            p.ch = _lo_quote
        case '}':
            return
    }
    p.ch = _l_nil
    for {
        switch p.ch {
            case _lo_quote:
                depth++
            case _lc_quote:
                depth--
                if depth == 0 {
                    //up to caller to step over last }
                    p.escm = _reg
                    return
                }
            case _eof:
                SyntaxError("{ without }")
        }
        p._next()
    }
}

func (p *_parser) _parse_quote() *sNode {
    var q Quote
    p._rquote(true)
    out := p._read_out()
    p._next()
    if len(out) == 0 {
        //got {}
        q = Noop
    } else {
        q = &protected_quote{&quote{false, nil, out}}
    }
    return &sNode{synQuote, q, nil}
}

func (p *_parser) _parse_string() *sNode {
    p.escm = _str
    p._next()
    if p.ch == _l_str {
        //got ""
        p._next()
        p.escm = _reg
        return &sNode{synLiteral, Null, nil}
    }
    p.record = true
    for ; p.ch != _l_str; p._next() {
        if p.ch == _eof {
            SyntaxError("\" without \"")
        }
    }
    val := &sNode{synLiteral, intern(p._read_out()), nil}
    p.escm = _reg
    p._next()
    return val
}

func (p *_parser) _parse_line(clause bool) *sNode {
    var head, node *sNode
    join := func(n *sNode) {
        if head != nil {
            node.next = n
            node = n
        } else {
            head = n
            node = head
        }
    }
    for ; p.ch == _l_space; p._next() {} //skip leading ws
    if clause && p.ch == _lc_clause {
        SyntaxError("[] invalid. Use {} for no-op")
    }
    switch p.ch {
        case _eol, _eof:
            //just a blank line
            if clause {
                SyntaxError("[ without ]")
            }
            p._next() //only needed for eol, no effect on eof
            return nil
    }
    //check for comments. comments have to have balanced {}
    //in case they're in an unparsed quote
    if !clause && p.ch == _l_comment {
        for ; p.ch != _eol; p._next() {
            switch p.ch {
                case _eof:
                    return nil
                case _lo_quote:
                    //we don't _next because loop takes care of it
                    p._rquote(false)
            }
        }
    }
    for {
        for ; p.ch == _l_space; p._next() {} //skip ws
        switch p.ch {
            case _lc_quote:
                SyntaxError("} before {")
            case _eof, _eol:
                if clause {
                    SyntaxError("[ without ]")
                }
                p._next()
                return head
            case _lc_clause:
                if !clause {
                    SyntaxError("] before [")
                }
                p._next()
                return &sNode{synClause, head, nil}
            case _lo_clause:
                p._next()
                join(p._parse_line(true))
            case _l_str:
                join(p._parse_string())
            case _lo_quote:
                join(p._parse_quote())
            default:
                join(p._parse_word())
        }
    }
    panic("parse line in impossible state")//Issue 65
}

func parse(in io.Reader) *command {
    var head, tail *command
    var n *sNode
    p := new(_parser)
    p.src = in
    p.cur = make([]byte, 1, 1)
    p.buf = newBuf(0)
    p.escm = _reg
    p._next() //prime l'pump
    for p.ch != _eof {
        if n = p._parse_line(false); n != nil {
            node := &command{n, nil}
            if head != nil {
                tail.next = node
                tail = tail.next
            } else {
                head = node
                tail = head
            }
        }
    }
    parse_trace("quotation has parsed to", head)
    return head
}
