package commands

import "gelo"

func Puts(vm *gelo.VM, args *gelo.List, _ uint) gelo.Word {
	if args == nil {
		return gelo.Null
	}
	vm.API.Send(args)
	return args
}

func Gets(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 0 {
		gelo.ArgumentError(vm, "gets", "", args)
	}
	return vm.API.Recv()
}

var IOCommands = map[string]interface{}{
	"puts": Puts,
	"gets": Gets,
}
