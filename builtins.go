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
			case Error:
				ret = t
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

	n, _ := NewNumberFromGo(spawned.id)
	return n
}

var BI_defer = &defert{}

var Core = map[string]interface{}{
	"defer":     BI_defer,
	"eval":      BI_eval,
	"safe-eval": BI_safe_eval,
	"go":        BI_go,
}
