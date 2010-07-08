package gelo

import (
	"fmt"
	"bytes"
)

//hopefully one day we also have name, lineno, etc
type _error struct {
	from uint32
	msg  string
}

//we do not let _errSystem become a gelo.Error but we do want to piggyback
//on the formatting. This is an unexported type because this kind of error is
//used for serious bugs that if allowed to continue would only wreak
//unimaginable havoc, so we want our SOS to escape even if we have to blow up
//the host application to do so. Should be caught by anything looking for
//os.Error and make it to a log if not the console.
type _errSystem struct {
	_error
}

//This error type isn't meant to be caught by this program or the host program
//it signals a static misuse of Gelo that needs to be corrected. Such as trying
//to run a VM without a program or use a VM that has been killed
type HostProgrammerError struct {
	_error
}

type ErrSyntax struct {
	_error
}

type ErrRuntime struct {
	_error
}

//Predefined kinds of errors

func ProgrammerError(vm *VM, s ...interface{}) {
	panic(HostProgrammerError{_make_error(vm, s)})
}

//note a system error is not a word and should only be called if the impossible
//to recover from occurs
func SystemError(vm *VM, s ...interface{}) {
	panic(_errSystem{_make_errorM(vm, s, `
If you are reading this, you have discovered a bug in Gelo, a  and not the
program that you are using.
Please see if this has been reported at:

	http://code.google.com/p/gelo/issues/list

and if not report the error there, so that we may fix it, before notifying the
owners of the application that you are using of the error.

Thank you for your time and the Gelo team deeply apologizes for the inconvience.
`)})
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

func killed(vm *VM) Error {
	return ErrRuntime{_make_error(vm, []interface{}{"VM killed"})}
}

//common methods on error

func (e _error) From() uint32 {
	return e.from
}

func (e _error) String() string {

	return e.msg
}

func (e _error) Message() string {
	return e.msg
}


//syntax errors

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


//Runtime Errors

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


//Implementation details

func _make_errorM(vm *VM, s ...interface{}) _error {
	return _make_error(vm, s)
}

func _make_error(vm *VM, s []interface{}) _error {
	var id uint32
	if vm != nil {
		id = vm.ProcID()
	}
	return _error{id, _format(s)}
}

func _format(all ...interface{}) string {
	return _format_slice(all)
}

func _format_slice(all []interface{}) string {
	buf := newBuf(0)
	buf.Write(_format1(all[0]))
	for _, v := range all[1:] {
		buf.WriteString(" ")
		buf.Write(_format1(v))
	}
	return buf.String()
}

func _format1(item interface{}) (ret []byte) {
	switch t := item.(type) {
	case nil:
		ret = []byte("NIL")
	case string:
		if len(t) == 0 {
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
		ret = []byte(_format_slice(t))
		/*buf := newBuf(0)
		buf.WriteString("{")
		if len(t) != 0 {
			buf.Write(_format1(t[0]))
			for _, v := range t[1:] {
				buf.WriteString(" ")
				buf.Write(_format1(v))
			}
		}
		buf.WriteString("}")
		ret = buf.Bytes()*/
	default:
		ret = []byte(fmt.Sprint(t))
	}
	return
}
