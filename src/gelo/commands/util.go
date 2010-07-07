package commands

import (
	"gelo"
	"math"
)

func possiblyInvokable(w gelo.Word) bool {
	_, ok1 := w.(gelo.Symbol) //can't know it will always point to an invokable
	_, ok2 := w.(gelo.Quote)
	_, ok3 := w.(gelo.Alien)
	return ok1 || ok2 || ok3
}

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