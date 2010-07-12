package commands

import (
	"gelo"
	"math"
)

func IndexError(vm *gelo.VM, idx int, w gelo.Word) {
	gelo.RuntimeError(vm, "Invalid index", idx, "in", w)
}

func ToIdx(vm *gelo.VM, w gelo.Word, length int) int {
	n := vm.API.NumberOrElse(w)
	j, ok := n.Int()
	if !ok || math.MaxInt32 < j {
		gelo.TypeMismatch(vm, "32-bit integer", "float")
	}
	i := int(j)
	if (i < 0 && i < -length) || (i >= 0 && i >= length) {
		IndexError(vm, i, n)
	}
	return (length + i) % length
}

func Aggregate(items map[string]interface{}) gelo.Alien {
	Map := make(map[string]gelo.Word)
	for k, v := range items {
		Map[k] = gelo.Convert(v)
	}
	d := gelo.NewDictFrom(Map)
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		if ac == 0 {
			return d
		}
		item, there := Map[args.Value.Ser().String()]
		if !there {
			//XXX this error message could be comprehensible in theory
			//will fix itself when vm contains name, lineno, etc
			gelo.ArgumentError(vm, "<an aggregate>", "command args*", args.Next)
		}
		return vm.API.TailInvokeCmd(item, args.Next)
	})
}
