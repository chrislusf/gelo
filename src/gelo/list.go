package gelo

var EmptyList *List = nil

func NewList(s ...Word) *List {
	return NewListFrom(s)
}

func NewListFrom(s []Word) *List {
	if len(s) == 0 {
		return EmptyList
	}
	head := &List{s[0], nil}
	tail := head
	for _, v := range s[1:] {
		tail.Next = &List{v.Copy(), nil}
		tail = tail.Next
	}
	return head
}

func NewListFromGo(s []interface{}) *List {
	if len(s) == 0 {
		return EmptyList
	}
	head := &List{Convert(s[0]), nil}
	if len(s) == 1 {
		return head
	}
	tail := head
	for _, v := range s[1:] {
		tail.Next = &List{Convert(v), nil}
		tail = tail.Next
	}
	return head
}

//Takes a space separated string of values, unescapes;
//set enc to true if the outer {} is included in the string
func UnserializeList(ser []byte, enc bool) (*List, bool) {
	var head, tail *List
	pos, ok := SlurpWS(ser, 0), false
	var s []byte
	if enc {
		if ser[pos] != '{' {
			return nil, false
		}
		pos++
	}
	for pos < len(ser) {
		s, pos, ok = UnescapeItem(ser, SlurpWS(ser, pos))
		if !ok || (pos < len(ser) && ser[pos] != ' ') {
			return nil, false
		}
		pos = SlurpWS(ser, pos)
		sym := BytesToSym(s)
		if head != nil {
			tail.Next = &List{sym, nil}
			tail = tail.Next
		} else {
			head = &List{sym, nil}
			tail = head
		}
	}
	if enc && ser[pos] != '}' {
		return nil, false
	}
	pos = SlurpWS(ser, pos)
	if pos != len(ser) {
		return nil, false
	}
	return head, true
}

func UnserializeListFrom(w Word) (*List, bool) {
	enc := false
	var ser []byte
	switch t := w.(type) {
	default:
		return nil, false
	case Symbol:
		enc = true
		ser = t.Bytes()
	case Quote:
		ser = t.Ser().Bytes()
	}
	l, ok := UnserializeList(ser, enc)
	if !ok {
		return nil, false
	}
	return l, true
}

//If w is a list then return, otherwise return a singleton list wrapping w
func AsList(w Word) *List {
	if l, ok := w.(*List); ok {
		return l
	}
	return &List{w, nil}
}

func (l *List) Slice() []Word {
	if l == nil {
		return nil
	}
	ret := make([]Word, 0, 16)
	for i := 0; l != nil; l = l.Next {
		if i == cap(ret) {
			new := make([]Word, len(ret), 2*cap(ret))
			copy(new, ret)
			ret = new
		}
		ret = ret[0 : i+1]
		ret[i] = l.Value
		i++
	}
	return ret
}

func (l *List) Len() (count int) {
	for ; l != nil; l = l.Next {
		count++
	}
	return count
}

func (l *List) Map(f func(Word) Word) *List {
	if l == nil {
		return EmptyList
	}
	head := &List{f(l.Value), nil}
	l = l.Next
	for tail := head; l != nil; l, tail = l.Next, tail.Next {
		tail.Next = &List{f(l.Value), nil}
	}
	return head
}

// If len==0 return Null
// If len==1 return f(head of list)
// Otherwise return list.Map(f)
func (l *List) MapOrApply(f func(Word) Word) Word {
	if l == nil {
		return Null
	}
	if l.Next == nil {
		return f(l.Value)
	}
	return l.Map(f)
}

func (_ *List) Type() Symbol { return interns("*LIST*") }

func (l *List) Ser() Symbol {
	var bytes []byte
	buf := newBuf(0)
	buf.WriteString("{")
	if l != nil {
		for ; l.Next != nil; l = l.Next {
			bytes = EscapeItem(l.Value.Ser().Bytes())
			buf.Write(bytes)
			buf.WriteString(" ")
		}
		bytes = EscapeItem(l.Value.Ser().Bytes())
		buf.Write(bytes)
	}
	buf.WriteString("}")
	return buf.Symbol()
}

func (l *List) Copy() Word {
	//WARNING/TODO does not detect cycles
	var head, tail *List
	for ; l != nil; l = l.Next {
		if head != nil {
			tail.Next = &List{l.Value, nil}
			tail = tail.Next
		} else {
			head = &List{l.Value, nil}
			tail = head
		}
	}
	return head
}

func (l *List) DeepCopy() Word {
	//WARNING/TODO does not detect cycles
	var head, tail *List
	for ; l != nil; l = l.Next {
		if head != nil {
			tail.Next = &List{l.Value.DeepCopy(), nil}
			tail = tail.Next
		} else {
			head = &List{l.Value.DeepCopy(), nil}
			tail = head
		}
	}
	return head
}

func (l *List) Equals(w Word) bool {
	ol, ok := w.(*List)
	if !ok {
		return false
	}
	for {
		if ol == nil {
			return l != nil
		}
		if l == nil {
			return ol != nil
		}
		if !l.Value.Equals(ol.Value) {
			return false
		}
		l, ol = l.Next, ol.Next
	}
	return true
}
