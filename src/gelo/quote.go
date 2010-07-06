package gelo

import "bytes"

var Noop = &protected_quote{&quote{false, nil, []byte("")}}

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
                if _, test := x.(ErrSyntax); test {
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
    literal   bool    //false until proven otherwise
    code     *command
    source  []byte
}

func (q *quote) unprotect() *quote { return q }

func (q *quote) Equals(w Word) bool {
    oq, ok := w.(Quote)
    if !ok {
        return false
    }
    return bytes.Equal(q.source, oq.unprotect().source)
}

func (q *quote) Ser() Symbol { return BytesToSym(dup(q.source)) }

func (_ *quote) Type() Symbol { return interns("*QUOTE*") }

//quotes are immutable
func (q *quote) Copy() Word { return q }
func (q *quote) DeepCopy() Word { return q }


type protected_quote struct {
    protectee *quote
}

func (q *protected_quote) unprotect() *quote { return q.protectee }

func (q *protected_quote) Equals(w Word) bool {
    oq, ok := w.(Quote)
    if !ok {
        return false
    }
    return bytes.Equal(q.unprotect().source, oq.unprotect().source)
}

func (q *protected_quote) Ser() Symbol { return q.protectee.Ser() }

func (_ *protected_quote) Type() Symbol { return interns("*QUOTE*") }

func (q *protected_quote) Copy() Word { return q }
func (q *protected_quote) DeepCopy() Word { return q }
