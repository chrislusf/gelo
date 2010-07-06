package gelo

import (
    "fmt"
    "bytes"
)

//note a system error is not a word and should only be called if the impossible
//to recover from occurs
func SystemError(vm *VM, s ...interface{}) {
    panic(ErrSystem(_make_error(vm, s)))
}

func SyntaxError(s ...interface{}) {
    panic(ErrSyntax{_make_error(nil, s)})
}

func VariableUndefined(vm *VM, name interface{}) {
    panic(ErrRuntime{_make_errorM(vm, "Undefined variable:", name)})
}

func RuntimeError(vm *VM, s ...interface{}) {
    panic(ErrRuntime{_make_error(vm, s)})
}

func TypeMismatch(vm *VM, exp, got interface{}) {
    panic(ErrRuntime{
        _make_errorM(vm, "Type mismatch. Expected:", exp, "Got:", got),
    })
}

//TODO get name from VM when it saves it
func ArgumentError(vm *VM, name, spec, args interface{}) {
    if args == nil {
        args = "no arguments"
    }
    panic(ErrRuntime{
     _make_errorM(vm, "Illegal arguments.", name, "expected:", spec, "Got:",
     args),
    })
}


func _make_errorM(vm *VM, s ...interface{}) _error {
    return _make_error(vm, s)
}

func _make_error(vm *VM, s []interface{}) _error {
    var id vm_id 
    if vm != nil {
        id = vm.ProcID()
    }
    return _error{id, _format(s)}
}

//hopefully one day we also have name, lineno, etc
type _error struct {
    from vm_id
    msg  string
}

func (e _error) From() vm_id {
    return e.from
}

func (e _error) String() string {
    return e.msg
}

type ErrSystem  _error //not a word

type ErrSyntax  struct { _error }

func (self ErrSyntax) Type() Symbol {
    return interns("*SYNTAX-ERROR*")
}

func (self ErrSyntax) Ser() Symbol {
    return Convert("Syntax error: " + self.String()).(Symbol)
}

func (self ErrSyntax) Equals(w Word) bool {
    e, ok := w.(ErrSyntax)
    if !ok {
        return false
    }
    return bytes.Equal([]byte(self.msg), []byte(e.msg))
}

func (self ErrSyntax) Copy() Word { return self }

func (self ErrSyntax) DeepCopy() Word { return self }


type ErrRuntime struct { _error }

func (self ErrRuntime) Type() Symbol {
    return interns("*RUNTIME-ERROR*")
}

func (self ErrRuntime) Ser() Symbol {
    return StrToSym("Runtime error: " + self.String())
}

func (self ErrRuntime) Equals(w Word) bool {
    e, ok := w.(ErrRuntime)
    if !ok {
        return false
    }
    return bytes.Equal([]byte(self.msg), []byte(e.msg))
}

func (self ErrRuntime) Copy() Word { return self }

func (self ErrRuntime) DeepCopy() Word { return self }



func _format1(item interface{}) (ret []byte) {
    switch t := item.(type) {
        case nil:
            ret = []byte("NIL")
        case string:
            if len(t)==0 {
                ret = []byte("\"\"")
            } else {
                ret = []byte(t)
            }
        case []byte:
            if len(t) == 0 {
                ret = []byte("\"\"")
            } else {
                ret = t
            }
        case *List:
            buf := newBuf(0)
            for ; t != nil; t = t.Next {
                buf.Write(_format1(t.Value))
                buf.WriteString(" ")
            }
            ret = buf.Bytes()
        case Symbol:
            ret = t.Bytes()
            if len(ret) == 0 {
                ret = []byte("<<the null string>>")
            }
        case Word:
            ret = t.Ser().Bytes()
        case []interface{}:
            buf := newBuf(0)
            buf.WriteString("{")
            if len(t) != 0 {
                buf.Write(_format1(t[0]))
                for _, v := range t[1:] {
                    buf.WriteString(" ")
                    buf.Write(_format1(v))
                }
            }
            buf.WriteString("}")
            ret = buf.Bytes()
        default:
            ret = []byte(fmt.Sprint(t))
    }
    return
}

func _format(all ...interface{}) string {
    buf := newBuf(0)
    buf.Write(_format1(all[0]))
    for _, v := range all[1:] {
        buf.WriteString(" ")
        buf.Write(_format1(v))
    }
    return buf.String()
}
