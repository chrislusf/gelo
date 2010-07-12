package gelo

import (
	"bytes"
	"os"
)

type reader interface {
	Read([]byte) (int, os.Error)
}

type _recordingReader struct {
	buf *bytes.Buffer
	src reader
}

func newRecordingReader(src reader) *_recordingReader {
	if rr, ok := src.(*_recordingReader); ok {
		return rr
	}
	return &_recordingReader{new(bytes.Buffer), src}
}

func (r *_recordingReader) Read(p []byte) (n int, err os.Error) {
	n, err = r.src.Read(p)
	if err != nil {
		return
	}
	r.buf.Write(p[0:n])
	return
}

func (r *_recordingReader) Bytes() []byte {
	return r.buf.Bytes()
}

type buffer struct {
	*bytes.Buffer
}

func newBuf(sz int) *buffer {
	if sz < 64 {
		return &buffer{new(bytes.Buffer)}
	}
	return &buffer{bytes.NewBuffer(make([]byte, 0, sz))}
}

func newBufFrom(s []byte) *buffer {
	return &buffer{bytes.NewBuffer(s)}
}

func newBufFromString(s string) *buffer {
	return &buffer{bytes.NewBufferString(s)}
}

func (b *buffer) WriteWord(w Word) {
	if w == nil { //XXX this is just for catching errors
		b.WriteString("NIL")
		return
	}
	if l, ok := w.(*List); ok && l == nil {
		b.WriteString("{}")
		return
	}
	b.Write(w.Ser().Bytes())
}

func (b *buffer) CopyBytes() []byte {
	out := make([]byte, b.Buffer.Len())
	copy(out, b.Buffer.Bytes())
	b.Buffer.Reset()
	return out
}

func (b *buffer) CopyString() string {
	return string(b.CopyBytes())
}

func (b *buffer) Symbol() Symbol {
	return BytesToSym(b.Bytes())
}

func (b *buffer) CopySymbol() Symbol {
	return BytesToSym(b.CopyBytes())
}

func (b *buffer) Bytes() []byte {
	out := b.Buffer.Bytes()
	b.Buffer.Reset()
	return out
}

func (b *buffer) String() string {
	if b.Len() == 0 {
		return ""
	}
	return string(b.Bytes())
}

func dup(s []byte) []byte {
	out := make([]byte, len(s))
	copy(out, s)
	return out
}

//used by list and dict's Ser methods
func EscapeItem(item []byte) []byte {
	var cur byte
	var out []byte
	is_str := false
	buf := newBuf(0)
	buf.WriteString("\"") //stripped if is_str is false at the end
	for pos := 0; pos < len(item); pos++ {
		cur = item[pos]
		switch cur {
		case ' ', '\t', '\f', '\n':
			is_str = true
		case '\\', '"', '{', '}':
			buf.WriteString("\\")
		}
		buf.WriteByte(cur)
	}
	if is_str || buf.Len() == 1 {
		buf.WriteString("\"")
		out = buf.Bytes()
	} else {
		//strip the initial "
		out = buf.Bytes()[1:]
	}
	return out
}

func UnescapeItem(item []byte, pos int) ([]byte, int, bool) {
	str := false
	if item[pos] == '"' {
		str = true
		pos++
		if pos >= len(item) {
			return nil, 0, false
		}
		if item[pos] == '"' {
			//Null
			return []byte(""), pos + 1, true
		}
	}
	buf := newBuf(0)
	for ; pos < len(item); pos++ {
		if item[pos] == '\\' {
			pos++
			if pos >= len(item) {
				return nil, 0, false
			}
			buf.WriteByte(item[pos])
			pos++
			continue
		}
		if str && item[pos] == '"' {
			pos++
			break
		}
		if !str && item[pos] == ' ' {
			break
		}
		if item[pos] == '}' {
			if !str {
				break
			} else {
				//invalid encoding
				return nil, 0, false
			}
		}
		buf.WriteByte(item[pos])
	}
	return buf.Bytes(), pos, true
}

func SlurpWS(s []byte, pos int) int {
	if pos >= len(s) {
		return pos
	}
	for {
		switch s[pos] {
		case ' ', '\n', '\f', '\t':
			pos++
			if pos >= len(s) {
				return pos
			}
		default:
			return pos
		}
	}
	panic("SlurpWS in impossible state") //Issue 65
}

func Convert(item interface{}) Word {
	var word Word
	word, ok := NewNumberFromGo(item) //easier to check this first
	if !ok {
		switch t := item.(type) {
		default:
			programmerError(nil, "Convert given unknown type")
		case nil:
			word = Null
		case func(*VM, *List, uint) Word:
			word = Alien(t)
		case Word:
			word = t
		case *defert:
			word = t
		case bool:
			word = ToBool(t)
		case string:
			word = interns(t)
		case []byte:
			word = intern(t)
		case []int:
			word = interns(string(t))
		case []string:
			if len(t) == 0 {
				word = EmptyList
			} else {
				l := &List{interns(t[0]), nil}
				word = Word(l)
				if len(t) > 1 {
					for _, val := range t[1:] {
						l.Next = &List{interns(val), nil}
						l = l.Next
					}
				}
			}
		case []Word:
			word = NewListFrom(t)
		case []interface{}:
			if len(t) == 0 {
				word = EmptyList
			} else {
				l := &List{Convert(t[0]), nil}
				word = l
				if len(t) > 1 {
					for _, val := range t[1:] {
						l.Next = &List{Convert(val), nil}
						l = l.Next
					}
				}
			}
		case map[string]interface{}:
			tmp := make(map[string]Word)
			for k, v := range t {
				tmp[k] = Convert(v)
			}
			word = &Dict{rep: tmp}
		case map[string]Word:
			word = NewDictFrom(t)
		}
	}
	return word
}
