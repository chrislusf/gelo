package commands

import "code.google.com/p/gelo"

func Die(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		args = gelo.AsList(gelo.StrToSym("die"))
	}
	gelo.RuntimeError(vm, args)
	return gelo.Null //Issue 65
}

func SyntaxError(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "SyntaxError", "error-msg+", "")
	}
	//TODO should get a line-no, etc to allow creation of good error message
	gelo.SyntaxError(args)
	panic("Issue 65")
}

func TypeMismatchError(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "TypeMismatchError",
			"expected-type recieved-type", args)
	}
	gelo.TypeMismatch(vm, args.Value.Ser(), args.Next.Value.Ser())
	panic("Issue 65")
}

func ArgumentError(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "ArgumentError", "name arg-spec args", args)
	}
	gelo.ArgumentError(vm, args.Value.Ser(), args.Next.Value.Ser(),
		args.Next.Next.Value.Ser())
	panic("Issue 65")
}

var ErrorCommands = map[string]interface{}{
	"die":               Die,
	"SyntaxError":       SyntaxError,
	"TypeMismatchError": TypeMismatchError,
	"ArgumentError":     ArgumentError,
}
