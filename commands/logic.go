package commands

import "code.google.com/p/gelo"

func And(vm *gelo.VM, args *gelo.List, ac uint) (ret gelo.Word) {
	switch ac {
	case 0:
		return gelo.True
	case 1:
		return args.Value
	}
	for ; args != nil; args = args.Next {
		ret = args.Value
		if b, ok := ret.(gelo.Bool); ok {
			if !b.True() {
				return gelo.False
			}
		}
	}
	return ret
}

func Or(vm *gelo.VM, args *gelo.List, ac uint) (ret gelo.Word) {
	switch ac {
	case 0:
		return gelo.False
	case 1:
		return args.Value
	}
	for ; args != nil; args = args.Next {
		ret = args.Value
		if b, ok := args.Value.(gelo.Bool); ok {
			if b.True() {
				return ret
			}
		} else {
			return ret
		}
	}
	return gelo.False
}

// Unlike And and Or, Not only works on bools
func Not(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return gelo.ToBool(!vm.API.BoolOrElse(w).True())
	})
}

func Equals(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	switch ac {
	case 0, 1:
		return gelo.True
	case 2:
		return gelo.ToBool(args.Value.Equals(args.Next.Value))
	}
	head := args.Value
	for args = args.Next; args != nil; args = args.Next {
		if !args.Value.Equals(head) {
			return gelo.False
		}
	}
	return gelo.True
}

func Not_equals(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	switch ac {
	case 0, 1:
		return gelo.False
	case 2:
		return gelo.ToBool(!args.Value.Equals(args.Next.Value))
	}
	head := args.Value
	for args = args.Next; args != nil; args = args.Next {
		if args.Value.Equals(head) {
			return gelo.False
		}
	}
	return gelo.True
}

var LogicCommands = map[string]interface{}{
	"and": And,
	"or":  Or,
	"not": Not,
	"=":   Equals,
	"/=":  Not_equals,
}
