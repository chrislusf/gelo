package gelo

import "os"

type Word interface {
    Type()          Symbol
    Ser()           Symbol
    Copy()          Word //these two should be in a MutableWord interface since
    DeepCopy()      Word //a lot of the intrinsic types are immutable
    Equals(Word)    bool
}

type Symbol interface {
    Word
    Bytes()    []byte
    Runes()    []int
    String()     string
    interned()   bool
}

type Quote interface {
    Word
    unprotect() *quote
}

type Port interface {
    Word
    Send(Word)
    Recv()      Word
    Close()
    Closed()    bool
}

type Error interface {
    os.Error
    Word
    From()  vm_id
}

type Bool bool

type Number struct {
    num   float64
    ser []byte
}

type List struct {
    Value   Word
    Next   *List
}

type Dict struct {
    rep   map[string]Word
    ser []byte
}

type Alien func(*VM, *List, uint) Word
func (_ Alien) Type() Symbol        { return interns("*ALIEN*") }
func (a Alien) Ser()  Symbol        { return a.Type() }
func (a Alien) Copy() Word          { return a }
func (a Alien) DeepCopy() Word      { return a }
func (a Alien) Equals(w Word) bool  {
    oa, ok := w.(Alien)
    if !ok {
        return false
    }
    return a == oa
}

//defined at the top of vm.go as it is a special internal tag
func (_ defert) Type() Symbol       { return interns("*DEFER*") }
func (d defert) Ser()  Symbol       { return d.Type() }
func (d defert) Copy() Word         { return d }
func (d defert) DeepCopy() Word     { return d }
func (_ defert) Equals(w Word) bool {
    _, ok := w.(defert)
    return ok
}
