package commands

import "gelo"

func Halt(vm *gelo.VM, args *gelo.List, _ uint) (_ gelo.Word) {
	vm.API.Halt(args)
	panic("halt in an impossible state") //Issue 65
}

func Id(_ *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	switch ac {
	case 0:
		return gelo.Null
	case 1:
		return args.Value
	}
	return args
}

//For each item, Value acts as the identity unless the item is a quote.
//If it is a quote attempt to invoke and return result if there were no errors
//If invocation fails for any reason Value returns the quote as a literal.
func Value(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "value", "items+", "")
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return vm.API.TailInvokeWordOrReturn(w)
	})
}

//Not the best name. Rewrites quote (ie rewrites $@[]) then returns
//a list of lists of symbols
func Partial_eval(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "partial-eval", "quote", args)
	}
	lists, ok := vm.API.PartialEval(vm.API.QuoteOrElse(args.Value))
	if !ok {
		gelo.TypeMismatch(vm, "code quote", "literal quote")
	}
	return lists
}

func Quote(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.Noop
	}
	if ac == 1 {
		return gelo.NewQuoteFrom(args.Value)
	}
	return gelo.NewQuoteFrom(args.Value.Ser())
}

func Invokablep(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		_, ok := vm.API.IsInvokable(w)
		return gelo.ToBool(ok)
	})
}

func InvokableOrId(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.Null
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		i, ok := vm.API.IsInvokable(w)
		if !ok {
			return w
		}
		return i
	})
}

func MakeInvokable(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.Noop
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		i, ok := vm.API.IsInvokable(w)
		if !ok {
			return gelo.Alien(
				func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
					if ac != 0 {
						gelo.ArgumentError(vm, w.Ser(), "", args)
					}
					return w
				})
		}
		return i
	})
}

var MiscCommands = map[string]interface{}{
	"halt":            Halt,
	"id":              Id,
	"value":           Value,
	"partial-eval":    Partial_eval,
	"Quote":           Quote,
	"invokable?":      Invokablep,
	"invokable-or-id": InvokableOrId,
	"force-invokable": MakeInvokable,
}
