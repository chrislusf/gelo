package gelo

import (
	"io"
	"sync"
)

const VERSION = "0.1.0 alpha"

type vm_id uint32

var external_id vm_id = 0 //called from Go code outside a VM
var _max_id vm_id = 0     //first VM is 1 since this gets inc'd
var _max_id_mutex sync.Mutex

type halt_control_code *List
type kill_control_code byte
type defert struct{}

type VM struct {
	API         *api
	Ns          *namespace_api
	cns, top    *namespace
	mux         *sync.RWMutex
	program     *quote
	io          Port
	running     bool
	id          vm_id
	kill_switch chan bool
	heritage    *_heritage
}

type _heritage struct {
	blacklist map[string]bool
	children  map[vm_id]chan bool
	parent    *VM
}

//a few boiler plate sanity checks to ensure that a destroyed VM
//isn't being operated upon externally
func (vm *VM) _sanity(msg string) {
	//if vm == nil we let the usual panic occur
	if vm.API == nil { //implies vm has been destroyed
		goto error
	} else if _, kill := <-vm.kill_switch; kill {
		vm.Destroy()
		goto error
	}
	return
error:
	programmerError(vm, "Dead VM attempted to "+msg)
}

// functions to create or destroy virtual machines

func _newVM(io Port) *VM {
	vm := &VM{io: io, mux: new(sync.RWMutex), kill_switch: make(chan bool)}
	//proxies
	vm.API = &api{vm}
	vm.Ns = &namespace_api{vm}
	_max_id_mutex.Lock()
	defer _max_id_mutex.Unlock()
	_max_id++
	vm.id = _max_id
	return vm
}

func NewVM(io Port) *VM {
	vm := _newVM(io)
	vm.cns = newNamespace(nil)
	vm.cns.set(argument_sym, Null)
	sys_trace("VM", vm.id, "created")
	return vm
}

func (vm *VM) Spawn() *VM {
	vm._sanity("spawn a child")
	vm2 := _newVM(vm.io)
	vm2.heritage = &_heritage{parent: vm}
	ns := newNamespace(vm.cns)
	vm2.top = vm.cns
	vm2.cns = ns
	//no parent
	if vm.heritage == nil {
		vm.heritage = &_heritage{children: make(map[vm_id]chan bool)}
	}
	//parent, no children
	if vm.heritage.children == nil { //has a parent but no children
		vm.heritage.children = make(map[vm_id]chan bool)
	}
	vm.heritage.children[vm2.id] = vm2.kill_switch
	vm2.heritage.parent = vm
	sys_trace("VM", vm2.id, "spawned from VM", vm.id)
	return vm2
}

func (vm *VM) Destroy() {
	if vm == nil {
		return
	}
	if vm.running {
		//therefore we are in a different goroutine
		vm.kill_switch <- true
		return
	}
	sys_trace("VM", vm.id, "destroyed")
	//either already dead or never had a parent
	if vm.heritage != nil {
		h := vm.heritage
		if h.parent != nil && h.parent.heritage != nil {
			//if we grab a reference before the field is set to nil in the
			//parent, it doesn't matter whether we delete our entry
			if pm := h.parent.heritage.children; pm != nil {
				pm[vm.id] = nil, false
			}
		}
		//if we spawned any VMs, kill them
		if h.children != nil {
			for _, child := range h.children {
				child <- true
			}
			h.children = nil
		} else {
			//If there were children we cannot free the ns pointers until
			//they are dead so they don't explode before they have a chance
			//to shut down, so we have to wait for the host to discard its
			//pointer to this VM for them to be collected.
			//If there were no children, however, we can safely discard them now
			vm.cns, vm.top = nil, nil
		}
		h.blacklist = nil
	}
	vm.kill_switch = nil
	vm.heritage = nil
	vm.io = nil
	if vm.API != nil {
		vm.API.vm = nil
	}
	vm.API = nil
	if vm.Ns != nil {
		vm.Ns.vm = nil
	}
	vm.Ns = nil
	vm.mux = nil
}

func Kill(vm *VM) {
	//if vm isn't nil but kill_switch is the vm has been destroyed but
	//the host is still holding on to a pointer
	if vm != nil {
		//grab a copy in case vm is destroyed in another thread
		//between the test and the send. Sending a kill to a destroyed VM
		//is safe.
		if kill_switch := vm.kill_switch; kill_switch != nil {
			sys_trace("VM", vm.id, "sent kill signal")
			kill_switch <- true
		}
	}
}

func (vm *VM) IsDead() bool {
	return vm == nil || vm.API == nil
}

func (vm *VM) IsRunning() bool {
	if vm == nil {
		return false
	}
	return vm.running
}

func (vm *VM) IsIdle() bool {
	return !vm.IsDead() && !vm.IsIdle()
}

//Change the io port of vm. It is not safe to call this while the vm is running.
func (vm *VM) Redirect(io Port) Port {
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	vm._sanity("redirect I/O port")
	p := vm.io
	vm.io = io
	return p
}

// commands to get information from a virtual machine

func (vm *VM) ProcID() uint32 {
	return uint32(vm.id)
}

//Register* -- add values to a VM

func (vm *VM) Register(name string, item interface{}) {
	vm._sanity("register an item")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	vm.cns.set(interns(name), Convert(item))
	vm.API.Trace("Registered:", name)
}

func (vm *VM) RegisterBundle(bundle map[string]interface{}) {
	for name, item := range bundle {
		vm.Register(name, item)
	}
}

func (vm *VM) RegisterBundles(bundles []map[string]interface{}) {
	for _, bundle := range bundles {
		vm.RegisterBundle(bundle)
	}
}

// Read* -- get values from a VM

func (vm *VM) ReadWord(name string) (Word, bool) {
	vm._sanity("read a word out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	return vm.cns.copyOut(name)
}

func (vm *VM) ReadString(name string) (string, bool) {
	vm._sanity("read a string out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	s, ok := vm.cns.copyOut(name)
	if !ok {
		return "", false
	}
	return s.Ser().String(), true
}

func (vm *VM) ReadBytes(name string) ([]byte, bool) {
	vm._sanity("read a byte string out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	s, ok := vm.cns.copyOut(name)
	if !ok {
		return nil, false
	}
	return s.Ser().Bytes(), true
}

func (vm *VM) ReadRunes(name string) ([]int, bool) {
	vm._sanity("read a byte string out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	s, ok := vm.cns.copyOut(name)
	if !ok {
		return nil, false
	}
	return s.Ser().Runes(), true
}

func (vm *VM) ReadBool(name string) (bool, bool) {
	vm._sanity("read a boolean out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	B, ok := vm.cns.copyOut(name)
	if !ok {
		return false, false
	}
	b, ok := B.(Bool)
	if !ok {
		return false, false
	}
	return b.True(), true
}

func (vm *VM) ReadMap(name string) (map[string]Word, bool) {
	vm._sanity("read a map out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	M, ok := vm.cns.copyOut(name)
	if !ok {
		return nil, false
	}
	m, ok := M.(*Dict)
	return map[string]Word(m.rep), true //no point in copying twice
}

func (vm *VM) ReadSlice(name string) ([]Word, bool) {
	vm._sanity("read a slice out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	S, ok := vm.cns.copyOut(name)
	if !ok {
		return nil, false
	}
	s, ok := S.(*List)
	if !ok {
		return nil, false
	}
	return s.Slice(), true //TODO copies twice
}

func (vm *VM) ReadQuote(name string) (Quote, bool) {
	vm._sanity("read a quote out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	Q, ok := vm.cns.copyOut(name)
	if !ok {
		return nil, false
	}
	q, ok := Q.(Quote)
	if !ok {
		return nil, false
	}
	return q, true
}

func (vm *VM) ReadPort(name string) (Port, bool) {
	vm._sanity("read a port out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	P, ok := vm.cns.copyOut(name)
	if !ok {
		return nil, false
	}
	p, ok := P.(Port)
	if !ok {
		return nil, false
	}
	return p, true
}

func (vm *VM) ReadChan(name string) (*Chan, bool) {
	vm._sanity("read a chan out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	C, ok := vm.cns.copyOut(name)
	if !ok {
		return nil, false
	}
	c, ok := C.(*Chan)
	if !ok {
		return nil, false
	}
	return c, true
}

func (vm *VM) ReadInt(name string) (int64, bool) {
	vm._sanity("read an integer out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	N, ok := vm.cns.copyOut(name)
	if !ok {
		return 0, false
	}
	n, ok := N.(*Number)
	if !ok {
		return 0, false
	}
	i, ok := n.Int()
	if !ok {
		return 0, false
	}
	return i, true
}

func (vm *VM) ReadFloat(name string) (float64, bool) {
	vm._sanity("read a number out")
	vm.mux.RLock()
	defer vm.mux.RUnlock()
	N, ok := vm.cns.copyOut(name)
	if !ok {
		return 0, false
	}
	n, ok := N.(*Number)
	if !ok {
		return 0, false
	}
	return n.Real(), true
}

// functions to run/change programs

func (vm *VM) SetProgram(q Quote) (err Error) {
	vm._sanity("set a new program")
	vm.mux.Lock()
	defer vm.mux.Unlock()
	iq := q.unprotect()
	_, ok := iq.fcode()
	if !ok {
		defer func() {
			if x := recover(); x != nil {
				if synerr, ok := x.(ErrSyntax); ok {
					err = synerr
					return
				}
				panic(x)
			}
		}()
		//force the error so we can return it
		err = force_synerr(vm, iq)
	}
	vm.program = iq
	return
}

func (vm *VM) GetProgram() Quote {
	if vm.mux != nil {
		vm.mux.RLock()
		defer vm.mux.RUnlock()
	}
	return &protected_quote{vm.program}
}

//Never call from a goroutine that doesn't own the VM
func (vm *VM) ParseProgram(in io.Reader) (err Error) {
	vm._sanity("parse and set a new program")
	vm.mux.Lock()
	defer vm.mux.Unlock()
	defer func() {
		if x := recover(); x != nil {
			if synerr, ok := x.(ErrSyntax); ok {
				err = synerr
				return
			}
			panic(x)
		}
	}()
	sys_trace("parsing")
	reader := newRecordingReader(in)
	code := parse(reader)
	vm.program = &quote{false, code, reader.Bytes()}
	return
}

//call from a different goroutine than the vm's and the outcome is undefined,
//if the VM is executing a program
func (vm *VM) Do(in string) (ret Word, err Error) {
	vm._sanity("execute: " + in)
	vm.mux.Lock()
	defer vm.mux.Unlock()
	defer func() {
		if x := recover(); x != nil {
			switch t := x.(type) {
			default:
				panic(x)
			case kill_control_code:
				sys_trace("VM", vm.id, "was killed")
				vm.Destroy()
				ret, err = nil, killed(vm)
			case halt_control_code:
				sys_trace("VM", vm.id, "halted")
				ret = (*List)(t)
			case ErrRuntime:
				//Syntax error would be in the source file
				//where Do is used so it must not be caught
				ret, err = nil, x.(Error)
			}
		}
	}()
	code := parse(newBufFromString(in))
	//we do this so a syntax error raised by the program can be caught but
	//a syntax error in 'in' is reported
	defer func() {
		if x := recover(); x != nil {
			if e, ok := x.(ErrSyntax); ok {
				ret, err = nil, e
			} else {
				panic(x)
			}
		}
		vm.running = false
	}()
	vm.running = true
	ret = vm.eval(code, nil)
	ret.DeepCopy()
	return
}

//Never call from a goroutine that doesn't own the VM
func (vm *VM) Exec(args interface{}) (ret Word, err Error) {
	vm._sanity("execute its program")
	vm.mux.Lock()
	defer vm.mux.Unlock()
	if vm.program == nil {
		programmerError(vm, "attempted to execute VM with no program")
	}
	defer func() {
		vm.running = false //true even if we error out before setting this true
		if x := recover(); x != nil {
			switch t := x.(type) {
			default:
				//either a _errSystem or bad programming
				//regardless, the system is now in a bad state so panic
				sys_trace("UNABLE TO RECOVER FROM PANIC")
				panic(x)
			case kill_control_code:
				sys_trace("VM", vm.id, "killed")
				vm.Destroy()
				ret, err = nil, killed(vm)
			case halt_control_code:
				sys_trace("VM", vm.id, "halted")
				ret = (*List)(t)
			case ErrSyntax, ErrRuntime:
				//there was a reasonable error, return it
				ret, err = nil, x.(Error)
			}
		} else {
			sys_trace("Program halted without error")
		}
	}()

	Args := EmptyList
	if args != nil {
		Args = AsList(Convert(args))
	}

	sys_trace("evaluating with arguments", Args)
	code, ok := vm.program.fcode()
	if !ok {
		//Somehow the program's quote was altered since it has been set
		systemError(vm, "The program has become corrupt")
	}

	vm.running = true //unset in defer handler
	ret = vm.eval(code, Args)
	ret = ret.DeepCopy()
	return
}

//same as ParseProgram followed by Exec
//Never call from a goroutine that doesn't own the VM
func (vm *VM) Run(in io.Reader, args interface{}) (ret Word, err Error) {
	if err = vm.ParseProgram(in); err != nil {
		return
	}
	return vm.Exec(args)
}
