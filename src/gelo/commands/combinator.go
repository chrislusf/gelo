package commands

import "gelo"

func _uninvokable(vm *gelo.VM, w gelo.Word) {
	gelo.RuntimeError(vm, w, "failed to invoke")
}

func Compose(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.Alien(Id)
	}
	if ac == 1 {
		if !possiblyInvokable(args.Value) {
			gelo.TypeMismatch(vm, "invokable", args.Value.Type())
		}
		return args.Value //TODO add a wrapper function to call it
	}
	//reverse and type check (as much as possible)
	var fs *gelo.List
	for ; args != nil; args = args.Next {
		v := args.Value
		if !possiblyInvokable(v) {
			gelo.TypeMismatch(vm, "invokable", v.Type())
		}
		fs = &gelo.List{v, fs}
	}
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, ac uint) (ret gelo.Word) {
		ret = args
		c := fs
		for ; c.Next != nil; c = c.Next {
			ret = gelo.AsList(vm.API.InvokeCmdOrElse(c.Value, ret.(*gelo.List)))
		}
		return vm.API.InvokeCmdOrElse(c.Value, ret.(*gelo.List))
	})
}

func Cleave(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac < 1 {
		gelo.ArgumentError(vm, "cleave", "cmds+", args)
	}
	var head, tail *gelo.List
	for ; args != nil; args = args.Next {
		v := args.Value
		if !possiblyInvokable(v) {
			gelo.TypeMismatch(vm, "invokable", v.Type())
		}
		if head != nil {
			tail.Next = &gelo.List{v, nil}
			tail = tail.Next
		} else {
			head = &gelo.List{v, nil}
			tail = head
		}
	}
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		h := gelo.NewList(vm.API.InvokeCmdOrElse(head.Value, args))
		t := h
		for c := head.Next; c != nil; c = c.Next {
			t.Next = &gelo.List{vm.API.InvokeCmdOrElse(c.Value, args), nil}
			t = t.Next
		}
		return h
	})
}

type _partial_kind byte

const (
	_pk_fill _partial_kind = iota
	_pk_hole
	_pk_slurp
)

type _partial struct {
	kind _partial_kind
	item gelo.Word
	next *_partial
}

func Partial(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	//TODO all of the argument errors in this function could be more informative
	if ac < 2 {
		gelo.ArgumentError(vm, "partial", "command [args*|'*|'X]+", args)
	}
	cmd, args := args.Value, args.Next
	if !possiblyInvokable(cmd) {
		gelo.TypeMismatch(vm, "invokable", cmd.Type())
	}
	a_star := false
	var val gelo.Word
	var kind _partial_kind
	var ph, pt *_partial
	for ; args != nil; args = args.Next {
		kind, val = _pk_fill, nil
		sym := vm.API.SymbolOrElse(args.Value)
		if gelo.StrEqualsSym("X", sym) {
			if a_star {
				gelo.ArgumentError(vm, "partial",
					"there will be no arguments left after *", args)
			}
			kind = _pk_hole
		} else if gelo.StrEqualsSym("*", sym) {
			if a_star {
				gelo.ArgumentError(vm, "partial",
					"only one * can be specfied", args)
			}
			kind, a_star = _pk_slurp, true
		} else {
			val = args.Value
		}
		if ph != nil {
			pt.next = &_partial{kind, val, nil}
			pt = pt.next
		} else {
			ph = &_partial{kind, val, nil}
			pt = ph
		}
	}
	pt = nil
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		real_args := &gelo.List{cmd, nil}
		tail := real_args
		for pls := ph; pls != nil; pls = pls.next {
			switch pls.kind {
			case _pk_fill:
				tail.Next = &gelo.List{pls.item, nil}
				tail = tail.Next
			case _pk_hole:
				if args == nil {
					gelo.ArgumentError(vm, "partial function",
						"requires at least one more argument", args)
				}
				tail.Next = &gelo.List{args.Value, nil}
				tail = tail.Next
				args = args.Next
			case _pk_slurp:
				if args != nil {
					tail.Next = args.Copy().(*gelo.List)
					for ; tail.Next != nil; tail = tail.Next {
					}
				}
				args = nil
			}
		}
		if args != nil {
			gelo.ArgumentError(vm, "partial function", "got too many arguments",
				args)
		}
		ret, err := vm.API.Invoke(real_args)
		if err != nil {
			panic(err)
		}
		return ret
	})
}

var CombinatorCommands = map[string]interface{}{
	"o":       Compose,
	"cleave":  Cleave,
	"partial": Partial,
}
