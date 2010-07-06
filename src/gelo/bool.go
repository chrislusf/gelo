package gelo

var True  = Bool(true)
var False = Bool(false)

func ToBool(b bool) Bool {
    if b {
        return True
    }
    return False
}

func (b Bool) True() bool { return bool(b) }

func (_ Bool) Type() Symbol { return interns("*BOOLEAN*") }

func (b Bool) Ser()  Symbol {
    if b {
        return interns("true")
    }
    return interns("false")
}
func (b Bool) Copy() Word {
    return b
}
func (b Bool) DeepCopy() Word {
    return b.Copy()
}
func (b Bool) Equals(w Word) bool {
    wb, ok := w.(Bool)
    if !ok {
        return false
    }
    return bool(b) == bool(wb)
}
