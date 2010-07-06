package gelo

import (
    "strconv"
    "math"
)

//Use this for ideal constants
func NewNumber(f float64) *Number {
    return &Number{num: f}
}

func NewNumberFrom(w Word) (*Number, bool) {
    return NewNumberFromString(w.Ser().String())
}

func NewNumberFromBytes(b []byte) (*Number, bool) {
    return NewNumberFromString(string(b))
}

func NewNumberFromString(s string) (*Number, bool) {
    num, err := strconv.Atof64(s)
    if err != nil {
        return nil, false
    }
    return &Number{num, []byte(s)}, true
}

func NewNumberFromGo(in interface{}) (*Number, bool) {
    var out   float64
    var ser []byte
    switch n := in.(type) {
        default:      return nil, false
        case nil:     return nil, false
        case *Number: return n, true
        case Word:    return NewNumberFrom(n)
        case []byte:  return NewNumberFromBytes(n)
        case string:  return NewNumberFromString(n)
        case float64: out = n
        case float32: out = float64(n)
        case int64:   out = float64(n)
        case int32:   out = float64(n)
        case int16:   out = float64(n)
        case int8:    out = float64(n)
        case int:     out = float64(n)
        case uint64:  out = float64(n)
        case uint32:  out = float64(n)
        case uint16:  out = float64(n)
        case uint8:   out = float64(n)
        case uint:    out = float64(n)
    }
    return &Number{out, ser}, true
}

func (n *Number) Real() float64 {
    return n.num
}

func (n *Number) Int() (int64, bool) {
    num := n.num
    if math.IsNaN(num) || math.Fmod(num, 1) != 0 || math.MaxInt64 < num {
        return 0, false
    }
    return int64(num), true
}

func (_ *Number) Type() Symbol {
    return interns("*NUMBER*")
}

func (n *Number) Ser() Symbol {
    if n.ser == nil {
        if i, ok := n.Int(); ok {
            n.ser = []byte(strconv.Itoa64(i))
        } else {
            n.ser = []byte(strconv.Ftoa64(n.num, 'g', -1))
        }
    }
    return BytesToSym(n.ser)
}

func (n *Number) Copy() Word {
    var ser []byte
    if n.ser != nil {
        ser = make([]byte, len(n.ser))
        copy(ser, n.ser)
    }
    return &Number{n.num, ser}
}

func (n *Number) DeepCopy() Word { return n.Copy() }

func (n *Number) Equals(w Word) bool {
    on, ok := w.(*Number)
    if !ok {
        return false
    }
    return on.num == n.num //XXX should check that they are within mach eps
}
