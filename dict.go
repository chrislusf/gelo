package gelo

func NewDict() *Dict {
	return &Dict{rep: make(map[string]Word)}
}

func NewDictFrom(m map[string]Word) *Dict {
	d := make(map[string]Word)
	for k, v := range m {
		d[k] = v.Copy()
	}
	return &Dict{rep: d}
}

func NewDictFromGo(m map[string]interface{}) *Dict {
	ret := NewDict()
	for k, v := range m {
		ret.rep[k] = Convert(v)
	}
	return ret
}

//Takes input string like "{k1 v1} {k2 v2} . . . {kN vN}", unescapes
//If enc = true, the above string is assumed to be wrapped in {}
func UnserializeDict(ser []byte, enc bool) (*Dict, bool) {
	d := make(map[string]Word)
	pos, ok := SlurpWS(ser, 0), false
	if !enc && pos >= len(ser) { //empty string->empty dict
		return NewDict(), true
	}
	if enc {
		if ser[pos] != '{' {
			return nil, false
		}
		pos = SlurpWS(ser, pos+1)
		//check for {}, if so return empty dict
		if pos < len(ser) && ser[pos] == '}' {
			pos = SlurpWS(ser, pos+1)
			if pos < len(ser) {
				return nil, false
			}
			return NewDict(), true
		}
	}
	//nonempty dict
	var key, val []byte
	for pos < len(ser) {
		//match {
		if ser[pos] != '{' {
			return nil, false
		}
		//key
		key, pos, ok = UnescapeItem(ser, SlurpWS(ser, pos+1))
		if !ok {
			return nil, false
		}
		//val
		val, pos, ok = UnescapeItem(ser, SlurpWS(ser, pos))
		if !ok {
			return nil, false
		}
		//match }
		pos = SlurpWS(ser, pos)
		if pos >= len(ser) || ser[pos] != '}' {
			return nil, false
		}
		pos = SlurpWS(ser, pos+1)
		//store
		d[string(key)] = BytesToSym(val)
		if enc && pos < len(ser) && ser[pos] == '}' {
			break
		}
	}
	if enc && pos < len(ser) && ser[pos] != '}' {
		return nil, false
	}
	pos = SlurpWS(ser, pos+1)
	//make sure there's no junk at the end
	if pos < len(ser) {
		return nil, false
	}
	return &Dict{rep: d}, true
}

func UnserializeDictFrom(w Word) (*Dict, bool) {
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
	d, ok := UnserializeDict(ser, enc)
	if !ok {
		return nil, false
	}
	return d, true
}

func (d *Dict) Map() map[string]Word {
	ret := make(map[string]Word)
	for k, v := range map[string]Word(d.rep) {
		ret[k] = v
	}
	return ret
}

func (d *Dict) Len() int {
	return len(d.rep)
}

//these methods sidestep the hashing restrictions on Go maps
func (d *Dict) Get(name Word) (w Word, ok bool) {
	return d.StrGet(stringof(name.Ser()))
}

func (d *Dict) StrGet(s string) (w Word, ok bool) {
	w, ok = d.rep[s]
	if !ok {
		w = Null
	}
	return
}

func (d *Dict) Set(name, value Word) {
	d.StrSet(stringof(name.Ser()), value)
}

func (d *Dict) StrSet(s string, w Word) {
	d.ser = nil
	d.rep[s] = w
}

func (d *Dict) Has(name Word) bool {
	return d.StrHas(stringof(name.Ser()))
}

func (d *Dict) StrHas(s string) bool {
	_, ok := d.rep[s]
	return ok
}

func (d *Dict) Del(name Word) {
	d.StrDel(stringof(name.Ser()))
}

func (d *Dict) StrDel(s string) {
	d.ser = nil
	delete(d.rep, s)
}

func (d *Dict) Ser() Symbol {
	if d.ser != nil {
		return BytesToSym(d.ser)
	}
	buf := newBuf(0)
	var bytes []byte
	buf.WriteString("{")
	for k, v := range d.rep {
		buf.WriteString("{")
		//key
		bytes = EscapeItem([]byte(k))
		buf.Write(bytes)
		buf.WriteString(" ")
		//value
		bytes = EscapeItem(v.Ser().Bytes())
		buf.Write(bytes)
		buf.WriteString("}")
	}
	buf.WriteString("}")
	d.ser = buf.Bytes()
	return BytesToSym(d.ser)
}

func (d *Dict) Equals(w Word) bool {
	od, ok := w.(*Dict)
	if !ok {
		return false
	}
	if len(d.rep) != len(od.rep) {
		return false
	}
	for k, v := range d.rep {
		ov, ok := od.rep[k]
		if !ok {
			return false
		}
		if !v.Equals(ov) {
			return false
		}
	}
	return true
}

func (d *Dict) Copy() Word {
	ret := NewDict()
	for k, v := range d.rep {
		ret.rep[k] = v
	}
	return ret
}

func (d *Dict) DeepCopy() Word {
	ret := NewDict()
	for k, v := range d.rep {
		ret.rep[k] = v.DeepCopy()
	}
	return ret
}

func (*Dict) Type() Symbol {
	return interns("*DICT*")
}
