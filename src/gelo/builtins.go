package gelo

//Commands for evaluation

func BI_eval(vm *VM, args *List, ac uint) (ret Word) {
	if ac == 0 {
		ArgumentError(vm, "eval", "code argument*", args)
	}
	defer func() {
		if x := recover(); x != nil {
			switch t := x.(type) {
			default:
				panic(x)
			case halt_control_code:
				ret = (*List)(t)
			case ErrSyntax, ErrRuntime:
				ret = x.(Error)
			}
		}
	}()
	//so we can't mess up the parent's namespace
	ns, top := vm.cns, vm.top
	defer func() { vm.cns, vm.top = ns, top }()
	vm.top = ns
	vm.cns = newNamespace(ns)
	//we use this rather than Invoke so we can catch halt's
	return vm.API.InvokeOrElse(args)
}

func _spawn(vm *VM, args *List) (*VM, *List) {
	head, tail := args.Value, args.Next
	if _, ok := head.(Quote); !ok {
		//if not a quote we conjure one around the args using dark powers
		head, tail = build_quote_from_list(args), EmptyList
	}
	spawned := vm.Spawn()
	if err := spawned.SetProgram(head.(Quote).unprotect()); err != nil {
		spawned.Destroy()
		panic(err) //going to be a syntax error
	}
	return spawned, tail
}

func BI_safe_eval(vm *VM, args *List, ac uint) Word {
	if ac == 0 {
		ArgumentError(vm, "safe-eval", "code argument*", args)
	}
	//TODO: add a "--with-env dict"
	spawned, rargs := _spawn(vm, args)
	defer spawned.Destroy()
	ret, err := spawned.Exec(rargs)
	if err != nil {
		return err
	}
	return ret
}

func _go_arg_err(vm *VM, args *List) {
	ArgumentError(vm, "go", "[--redirect port]? invokable argument*", args)
}

func BI_go(vm *VM, args *List, ac uint) Word {
	if ac == 0 {
		_go_arg_err(vm, args)
	}

	var port Port
	if StrEqualsSym("--redirect", args.Value.Ser()) {
		if ac < 3 {
			_go_arg_err(vm, args)
		}
		port = vm.API.PortOrElse(args.Next.Value)
		args = args.Next.Next
	}

	spawned, rargs := _spawn(vm, args)
	if port != nil {
		spawned.Redirect(port)
	}

	go func() {
		defer spawned.Destroy()
		vm.API.Trace("goroutine spawned")
		if _, err := spawned.Exec(rargs); err != nil {
			spawned.io.Send(err)
		}
	}()

	n, _ := NewNumberFromGo(uint(spawned.id))
	return n
}

var BI_defer = defert{}

var EvaluationCommands = map[string]interface{}{
	"defer":     BI_defer,
	"eval":      BI_eval,
	"safe-eval": BI_safe_eval,
	"go":        BI_go,
}

//Commands for variable manipulation
func BI_setx(vm *VM, args *List, ac uint) Word {
	if ac != 2 {
		ArgumentError(vm, "set!", "symbol value", args)
	}
	toset := args.Value.Ser()
	val := args.Next.Value
	vm.Ns.Set(toset, val)
	return val
}

func BI_updatex(vm *VM, args *List, ac uint) Word {
	if ac != 2 {
		ArgumentError(vm, "update!", "symbol value", args)
	}
	toreset := args.Value.Ser()
	val := args.Next.Value
	if !vm.Ns.Mutate(toreset, val) {
		VariableUndefined(vm, toreset)
	}
	return val
}

func BI_setp(vm *VM, args *List, ac uint) Word {
	if ac == 0 {
		ArgumentError(vm, "set?", "symbol+", args)
	}
	return args.MapOrApply(func(w Word) Word {
		return ToBool(vm.Ns.Has(w.Ser()))
	})
}

func BI_unsetx(vm *VM, args *List, ac uint) Word {
	if ac == 0 {
		ArgumentError(vm, "unset!", "symbol", args)
	}
	return args.MapOrApply(func(w Word) Word {
		todo := w.Ser()
		val, ok := vm.Ns.Del(todo)
		if !ok {
			VariableUndefined(vm, todo)
		}
		return val
	})
}

func BI_swapx(vm *VM, args *List, ac uint) Word {
	if ac != 2 {
		ArgumentError(vm, "swap!", "symbol symbol", args)
	}
	fst, snd := args.Value.Ser(), args.Next.Value.Ser()
	sndw, _, ok := vm.Ns.Swap(fst, snd) //don't need fstw
	if !ok {
		//we've lost our locks so this may produce nonsense
		fstb, sndb := vm.Ns.Has(fst), vm.Ns.Has(snd)
		if fstb && sndb {
			RuntimeError(vm, "Neither", fst, "nor", snd, "are defined")
		}
		if fstb {
			VariableUndefined(vm, fst)
		}
		if sndb {
			VariableUndefined(vm, snd)
		}
		RuntimeError(vm, "swap failed but both operands were defined")
	}
	vm.API.Trace("swapped the values of", fst, "and", snd)
	return sndw //returned because this is now the "closest"
}

//Commands for namespaces
func BI_NS_fork(vm *VM, args *List, ac uint) Word {
	if ac != 0 {
		ArgumentError(vm, "ns.fork", "", args)
	}
	vm.cns = newNamespace(vm.cns)
	vm.API.Trace("Child namespace created")
	return Null
}

func BI_NS_unfork(vm *VM, args *List, ac uint) Word {
	if ac != 0 {
		ArgumentError(vm, "ns.unfork", "", args)
	}
	d := vm.cns.dict
	up := vm.cns.up
	if up == nil || up == vm.top {
		RuntimeError(vm, "Fatal: Last namespace destroyed")
	}
	vm.cns = up
	vm.API.Trace("Namespace destroyed")
	return d
}

// wraps a quote in an alien that revivifies the calling namespace on each
// invocation.
func BI_NS_capture(vm *VM, args *List, ac uint) Word {
	if ac != 1 {
		ArgumentError(vm, "ns.capture", "code-quote", args)
	}
	dict, cmd := vm.cns.dict, vm.API.QuoteOrElse(args.Value)
	vm.API.Trace("current namespace captured")
	return Alien(func(vm *VM, args *List, _ uint) Word {
		ns := newNamespaceFrom(vm.cns, dict)
		vm.cns = ns
		defer func() {
			if vm.cns != ns {
				RuntimeError(vm,
					"Captured ns's code does not complete at captured ns")
			}
			vm.cns = ns.up
			ns.up, ns.dict = nil, nil
		}()
		vm.API.Trace("switching to captured namespace")
		return vm.API.InvokeCmdOrElse(cmd, args)
	})
}

func BI_NS_inject(vm *VM, args *List, ac uint) Word {
	if ac != 1 {
		ArgumentError(vm, "ns.inject", "dictionary", args)
	}
	d, ok := args.Value.(*Dict)
	if !ok {
		TypeMismatch(vm, "dictionary", args.Value)
	}
	ns := vm.cns
	ns.mux.Lock()
	defer ns.mux.Unlock()
	for k, v := range map[string]Word(d.rep) {
		ns.dict.Set(StrToSym(k), v)
	}
	return d
}

func _locals(vm *VM) map[string]Word {
	vm.cns.mux.RLock()
	defer vm.cns.mux.RUnlock()
	d := vm.cns.dict
	ret := make(map[string]Word)
	for k, v := range map[string]Word(d.rep) {
		ret[k] = v
	}
	return ret
}

func BI_NS_locals(vm *VM, args *List, ac uint) Word {
	if ac != 0 {
		ArgumentError(vm, "ns.locals", "", args)
	}
	return NewDictFrom(_locals(vm))
}

func BI_NS_globals(vm *VM, args *List, ac uint) Word {
	if ac != 0 {
		ArgumentError(vm, "ns.globals", "", args)
	}
	ret := _locals(vm)
	above := false
	for ns := vm.cns.up; ns != nil; ns = ns.up {
		if ns == vm.top {
			above = true
		}
		ns.mux.RLock()
		d := ns.dict
		for k, v := range map[string]Word(d.rep) {
			_, ok := ret[k]
			if !ok {
				if above {
					v = v.DeepCopy()
				}
				ret[k] = v
			}
		}
		ns.mux.RUnlock()
	}
	return NewDictFrom(ret)
}

func _go_up_n_levels(vm *VM, lvls int64) *namespace {
	ns := vm.cns
	for i := int64(0); i < lvls; i++ {
		ns = ns.up
		if ns == nil || ns == vm.top {
			RuntimeError(vm, "attempted to export to non-writeable namespace")
		}
	}
	return ns
}

func _parse_up_levels(vm *VM, up, n Word) (lvls int64, ok bool) {
	if StrEqualsSym("up", vm.API.SymbolOrElse(up)) {
		if lvls, ok = vm.API.NumberOrElse(n).Int(); ok {
			if lvls < 1 {
				s := "negative integer"
				if lvls == 0 {
					s = "zero"
				}
				TypeMismatch(vm, "nonzero positive integer", s)
			}
		} else {
			TypeMismatch(vm, "integer", "real")
		}
	}
	return
}

func _exportx_arg_err(vm *VM, args *List) {
	ArgumentError(vm, "export!", "[up levels]? name value", args)
}

func BI_exportx(vm *VM, args *List, ac uint) Word {
	if !(ac == 2 || ac == 4) {
		_exportx_arg_err(vm, args)
	}
	var ok bool
	lvls := int64(1)
	if ac == 4 {
		lvls, ok = _parse_up_levels(vm, args.Value, args.Next.Value)
		args = args.Next.Next
		if !ok {
			_exportx_arg_err(vm, args)
		}
	}
	wns := _go_up_n_levels(vm, lvls)
	name, value := vm.API.SymbolOrElse(args.Value), args.Next.Value
	wns.mux.Lock()
	defer wns.mux.Unlock()
	wns.dict.Set(name, value)
	return value
}

func BI_exportsx(vm *VM, args *List, ac uint) Word {
	if ac == 0 {
		ArgumentError(vm, "exports!", "[up levels]? name+", args)
	}
	lvls, ok := _parse_up_levels(vm, args.Value, args.Next.Value)
	if ok {
		args = args.Next.Next
	} else {
		lvls = 1
	}
	wns, rns := _go_up_n_levels(vm, lvls), vm.cns
	rns.mux.RLock()
	defer rns.mux.RUnlock()
	wns.mux.Lock()
	defer wns.mux.Unlock()
	for ; args != nil; args = args.Next {
		name := vm.API.SymbolOrElse(args.Value)
		value, ok := rns.dict.Get(name)
		if !ok {
			RuntimeError(vm, "Cannot export nonexistient symbol", name)
		}
		wns.dict.Set(name, value)
	}
	return Null
}

var VariableCommands = map[string]interface{}{
	"set!":     BI_setx,
	"update!":  BI_updatex,
	"set?":     BI_setp,
	"unset!":   BI_unsetx,
	"swap!":    BI_swapx,
	"export!":  BI_exportx,
	"exports!": BI_exportsx,
	"ns": Aggregate(map[string]interface{}{
		"fork":    BI_NS_fork,
		"unfork":  BI_NS_unfork,
		"capture": BI_NS_capture,
		"inject!": BI_NS_inject,
		"locals":  BI_NS_locals,
		"globals": BI_NS_globals,
	}),
}

var Values = map[string]interface{}{
	"true":  True,
	"false": False,
}

var CoreCommands = []map[string]interface{}{
	Values, EvaluationCommands, VariableCommands,
}
