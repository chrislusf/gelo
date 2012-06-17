package commands

import (
	"gelo"
	"math"
)

func NumberCon(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.NewNumber(0)
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		if n, ok := w.(*gelo.Number); ok {
			return n
		}
		if n, ok := gelo.NewNumberFromBytes(w.Ser().Bytes()); ok {
			return n
		}
		return gelo.False
	})
}

func Incrx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "incr!", "reference to number", args)
	}
	n, ok := vm.Ns.MutateBy(args.Value,
		func(w gelo.Word) (gelo.Word, bool) {
			n := vm.API.NumberOrElse(w)
			return gelo.NewNumber(n.Real() + 1), true
		})
	if !ok {
		gelo.VariableUndefined(vm, args.Value)
	}
	return n
}

func Decrx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "decr!", "reference to number", args)
	}
	n, ok := vm.Ns.MutateBy(args.Value,
		func(w gelo.Word) (gelo.Word, bool) {
			n := vm.API.NumberOrElse(w)
			return gelo.NewNumber(n.Real() - 1), true
		})
	if !ok {
		gelo.VariableUndefined(vm, args.Value)
	}
	return n
}

func Sum(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.NewNumber(0)
	}
	var acc float64
	for ; args != nil; args = args.Next {
		acc += vm.API.NumberOrElse(args.Value).Real()
	}
	return gelo.NewNumber(acc)
}

func Product(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.NewNumber(0)
	}
	var acc float64 = 1
	for ; args != nil; args = args.Next {
		acc *= vm.API.NumberOrElse(args.Value).Real()
	}
	return gelo.NewNumber(acc)
}

//Left-to-right
func Difference(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.NewNumber(0)
	}
	acc := vm.API.NumberOrElse(args.Value).Real()
	for args = args.Next; args != nil; args = args.Next {
		acc -= vm.API.NumberOrElse(args.Value).Real()
	}
	return gelo.NewNumber(acc)
}

//Left-to-right
func Quotient(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.NewNumber(0)
	}
	acc := vm.API.NumberOrElse(args.Value).Real()
	for args = args.Next; args != nil; args = args.Next {
		n := vm.API.NumberOrElse(args.Value).Real()
		if n == 0 {
			gelo.RuntimeError(vm, "Division by 0")
		}
		acc /= n
	}
	return gelo.NewNumber(acc)
}

func Mod(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "mod", "number base", args)
	}
	n := vm.API.NumberOrElse(args.Value).Real()
	m := vm.API.NumberOrElse(args.Next.Value).Real()
	return gelo.NewNumber(math.Mod(n, m))
}

func Lt(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	} else if ac == 1 {
		return gelo.True
	}
	last := vm.API.NumberOrElse(args.Value).Real()
	cur := last
	for args = args.Next; args != nil; args = args.Next {
		cur = vm.API.NumberOrElse(args.Value).Real()
		if last >= cur {
			return gelo.False
		}
		last = cur
	}
	return gelo.True
}

func Lte(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	} else if ac == 1 {
		return gelo.True
	}
	last := vm.API.NumberOrElse(args.Value).Real()
	cur := last
	for args = args.Next; args != nil; args = args.Next {
		cur = vm.API.NumberOrElse(args.Value).Real()
		if last > cur {
			return gelo.False
		}
		last = cur
	}
	return gelo.True
}

func Gt(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	} else if ac == 1 {
		return gelo.True
	}
	last := vm.API.NumberOrElse(args.Value).Real()
	cur := last
	for args = args.Next; args != nil; args = args.Next {
		cur = vm.API.NumberOrElse(args.Value).Real()
		if last <= cur {
			return gelo.False
		}
		last = cur
	}
	return gelo.True
}

func Gte(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	} else if ac == 1 {
		return gelo.True
	}
	last := vm.API.NumberOrElse(args.Value).Real()
	cur := last
	for args = args.Next; args != nil; args = args.Next {
		cur = vm.API.NumberOrElse(args.Value).Real()
		if last < cur {
			return gelo.False
		}
		last = cur
	}
	return gelo.True
}

func Min(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "min", "number+", args)
	}
	var min *gelo.Number
	var minf, candidate float64
	n := vm.API.NumberOrElse(args.Value)
	min = n
	minf = n.Real()
	for args = args.Next; args != nil; args = args.Next {
		n = vm.API.NumberOrElse(args.Value)
		candidate = n.Real()
		if candidate < minf {
			min = n
			minf = candidate
		}
	}
	return min
}

func Max(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "min", "number+", args)
	}
	var max *gelo.Number
	var maxf, candidate float64
	n := vm.API.NumberOrElse(args.Value)
	max = n
	maxf = n.Real()
	for args = args.Next; args != nil; args = args.Next {
		n = vm.API.NumberOrElse(args.Value)
		candidate = n.Real()
		if maxf < candidate {
			max = n
			maxf = candidate
		}
	}
	return max
}

func Neg(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.NewNumber(0)
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		n := vm.API.NumberOrElse(w).Real()
		return gelo.NewNumber(-n)
	})
}

func Abs(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.NewNumber(0)
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		n := vm.API.NumberOrElse(w).Real()
		return gelo.NewNumber(math.Abs(n))
	})
}

func Sgn(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "sgn", "number+", "")
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		n := vm.API.NumberOrElse(w).Real()
		switch {
		case n < 0:
			n = -1
		case n > 0:
			n = 1
		case n == 0:
			n = 0
		}
		return gelo.NewNumber(n)
	})
}

func Integerp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		n := vm.API.NumberOrElse(w)
		_, ok := n.Int()
		return gelo.ToBool(ok)
	})
}

func Positivep(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return gelo.ToBool(vm.API.NumberOrElse(w).Real() >= 0)
	})
}

func Negativep(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return gelo.ToBool(vm.API.NumberOrElse(w).Real() < 0)
	})
}

func NaNp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.False
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return gelo.ToBool(math.IsNaN(vm.API.NumberOrElse(w).Real()))
	})
}

func _infp_gen(s int) gelo.Alien {
	return func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		if ac == 0 {
			return gelo.False
		}
		return args.MapOrApply(func(w gelo.Word) gelo.Word {
			return gelo.ToBool(math.IsInf(vm.API.NumberOrElse(w).Real(), s))
		})
	}
}

var Infp = _infp_gen(0)
var PInfp = _infp_gen(1)
var NInfp = _infp_gen(-1)

var MathCommands = map[string]interface{}{
	"Number": NumberCon,
	"incr!":  Incrx,
	"decr!":  Decrx,
	"+":      Sum,
	"-":      Difference,
	"*":      Product,
	"div":    Quotient,
	"mod":    Mod,
	"<":      Lt,
	"<=":     Lte,
	">":      Gt,
	">=":     Gte,
	"min":    Min,
	"max":    Max,
	"abs":    Abs,
	"sgn":    Sgn,
	"neg":    Neg,
	//predicates
	"integer?":  Integerp,
	"positive?": Positivep,
	"negative?": Negativep,
	"NaN?":      NaNp,
	"Inf?":      Infp,
	"+Inf?":     PInfp,
	"-Inf?":     NInfp,
	//values
	"Inf":  math.Inf(0),
	"-Inf": math.Inf(-1),
	"Nan":  math.NaN(),
}
