package commands

import (
	"code.google.com/p/gelo"
	"code.google.com/p/gelo/extensions"
)

func Compose(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.Alien(Id) //defined in eval.go
	} else if ac == 1 {
		return vm.API.InvokableOrElse(args.Value)
	}
	//reverse and type check (as much as possible)
	builder := extensions.ListBuilder()
	defer builder.Destroy()
	for ; args != nil; args = args.Next {
		builder.PushFront(vm.API.InvokableOrElse(args.Value))
	}
	fs := builder.List()
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		c := fs
		var ret gelo.Word = args
		for ; c.Next != nil; c = c.Next {
			ret = gelo.AsList(vm.API.InvokeCmdOrElse(c.Value, ret.(*gelo.List)))
		}
		return vm.API.TailInvokeCmd(c.Value, ret.(*gelo.List))
	})
}

func Cleave(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac < 1 {
		gelo.ArgumentError(vm, "cleave", "cmds+", args)
	}
	builder := extensions.ListBuilder()
	defer builder.Destroy()
	for ; args != nil; args = args.Next {
		builder.Push(vm.API.InvokableOrElse(args.Value))
	}
	list := builder.List()
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		builder := extensions.ListBuilder()
		for c := list; c != nil; c = c.Next {
			builder.Push(vm.API.InvokeCmdOrElse(c.Value, args))
		}
		return builder.List()
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
	cmd, args := vm.API.InvokableOrElse(args.Value), args.Next

	a_star := false
	var val gelo.Word
	var kind _partial_kind
	var ph, pt *_partial
	for ; args != nil; args = args.Next {
		kind, val = _pk_fill, nil
		if sym, ok := args.Value.(gelo.Symbol); ok {
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
				val = sym
			}
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
		real_args := extensions.ListBuilder(cmd)
		for pls := ph; pls != nil; pls = pls.next {
			switch pls.kind {
			case _pk_fill:
				real_args.Push(pls.item)
			case _pk_hole:
				if args == nil {
					gelo.ArgumentError(vm, "partial function",
						"requires at least one more argument", args)
				}
				real_args.Push(args.Value)
				args = args.Next
			case _pk_slurp:
				real_args.Extend(args.Copy().(*gelo.List))
				args = nil
			}
		}
		if args != nil {
			gelo.ArgumentError(vm, "partial function", "got too many arguments",
				args)
		}
		return vm.API.InvokeOrElse(real_args.List())
	})
}

var CombinatorCommands = map[string]interface{}{
	"o":       Compose,
	"cleave":  Cleave,
	"partial": Partial,
}
