package commands

import "code.google.com/p/gelo"

func Copy(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "copy", "values+", "")
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return w.Copy()
	})
}

func Deep_copy(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "deep-copy", "values+", "")
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return w.DeepCopy()
	})
}

var CopyCommands = map[string]interface{}{
	"copy":      Copy,
	"deep-copy": Deep_copy,
}
