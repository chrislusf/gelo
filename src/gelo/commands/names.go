package commands

import "gelo"

func Setx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "set!", "name value", args)
	}
	val := args.Next.Value
	vm.Ns.Set(0, args.Value, val)
	return val
}

func Updatex(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "update!", "name value", args)
	}
	toreset := args.Value
	val := args.Next.Value
	if !vm.Ns.Mutate(toreset, val) {
		gelo.VariableUndefined(vm, toreset)
	}
	return val
}

func Setp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "set?", "name+", args)
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		_, ok := vm.Ns.Lookup(w)
		return gelo.ToBool(ok)
	})
}

func Unsetx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "unset!", "name+", args)
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		val, ok := vm.Ns.Del(w)
		if !ok {
			gelo.VariableUndefined(vm, w)
		}
		return val
	})
}

func Swapx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "swap!", "name name", args)
	}
	fst, snd := args.Value, args.Next.Value
	sndw, _, ok := vm.Ns.Swap(fst, snd) //don't need fstw
	if !ok {
		//we've lost our locks so this may produce nonsense
		_, fstb := vm.Ns.Lookup(fst)
		_, sndb := vm.Ns.Lookup(snd)
		if fstb && sndb {
			gelo.RuntimeError(vm, "Neither", fst, "nor", snd, "are defined")
		}
		if fstb {
			gelo.VariableUndefined(vm, fst)
		}
		if sndb {
			gelo.VariableUndefined(vm, snd)
		}
		gelo.RuntimeError(vm, "swap failed but both operands were defined")
	}
	return sndw //returned because this is now the "closest"
}

var _up_parser = MakeOrElseArgParser("[up levels]? args+")

func _export_parser(vm *gelo.VM, args *gelo.List) (int, *gelo.List) {
	Args := _up_parser(vm, args)
	levels, ok := Args["levels"]
	lvl := 1
	var lvl64 int64
	if ok {
		lvl64, ok = vm.API.NumberOrElse(levels).Int()
		if !ok || lvl < 0 {
			gelo.TypeMismatch(vm, "positive integer", "number")
		}
		lvl = int(lvl64)
	}
	return lvl, Args["args"].(*gelo.List)
}

func Exportx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	lvl := 1
	if ac == 4 {
		lvl, args = _export_parser(vm, args)
	} else if ac != 2 {
		gelo.ArgumentError(vm, "export!", "[up levels]? name value", args)
	}
	value := args.Next.Value
	if !vm.Ns.Set(lvl, args.Value, value) {
		gelo.RuntimeError(vm, "invalid namespace specified")
	}
	return value
}

func Exportsx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	lvls, args := _export_parser(vm, args)
	m := make(map[string]gelo.Word)
	for ; args != nil; args = args.Next {
		//Ns Ser's value anyway so no point in doing it twice
		k := args.Value.Ser()
		v := vm.Ns.LookupOrElse(k)
		m[k.String()] = v
	}
	d := gelo.NewDictFrom(m)
	if !vm.Ns.Inject(lvls, d) {
		gelo.RuntimeError(vm, "invalid namespace specified")
	}
	return d
}

//Commands for namespaces
func NS_fork(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 0 {
		gelo.ArgumentError(vm, "ns.fork", "", args)
	}
	vm.Ns.Fork(nil)
	vm.API.Trace("Child namespace created")
	return gelo.Null
}

func NS_unfork(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 0 {
		gelo.ArgumentError(vm, "ns.unfork", "", args)
	}
	if _, ok := vm.Ns.Unfork(); !ok {
		gelo.RuntimeError(vm, "Fatal: Last namespace destroyed")
	}
	vm.API.Trace("Namespace destroyed")
	return gelo.Null
}

func NS_inject(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "ns.inject", "dictionary", args)
	}
	d := vm.API.DictOrElse(args.Value)
	vm.Ns.Inject(0, d)
	return d
}

func NS_locals(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 0 {
		gelo.ArgumentError(vm, "ns.locals", "", args)
	}
	d := vm.Ns.Locals(0)
	return d
}

func NS_globals(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 0 {
		gelo.ArgumentError(vm, "ns.globals", "", args)
	}
	d := vm.Ns.Locals(-1)
	return d
}

func NS_capture(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "ns.capture", "code-quote", args)
	}
	cmd, dict := vm.API.QuoteOrElse(args.Value), vm.Ns.Locals(0)
	vm.API.Trace("current namespace captured")
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, _ uint) gelo.Word {
		vm.API.Trace("switching to captured namespace")
		vm.Ns.Fork(nil)
		defer vm.Ns.Unfork()
		vm.Ns.Inject(0, dict)
		//can't tail invoke because the ns would be unforked
		return vm.API.InvokeCmdOrElse(cmd, args)
	})
}

var VariableCommands = map[string]interface{}{
	"set!":     Setx,
	"update!":  Updatex,
	"set?":     Setp,
	"unset!":   Unsetx,
	"swap!":    Swapx,
	"export!":  Exportx,
	"exports!": Exportsx,
	"ns": Aggregate(map[string]interface{}{
		"fork":    NS_fork,
		"unfork":  NS_unfork,
		"capture": NS_capture,
		"inject!": NS_inject,
		"locals":  NS_locals,
		"globals": NS_globals,
	}),
}
