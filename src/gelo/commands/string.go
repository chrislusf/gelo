package commands

import (
	"gelo"
	"gelo/extensions"
	"bytes"
	"unicode"
	"utf8"
	"math"
)

func Chars(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac < 2 {
		gelo.ArgumentError(vm, "chars", "symbol indicies+", args)
	}
	str := vm.API.SymbolOrElse(args.Value).Bytes()
	if ac == 2 {
		i := ToIdx(vm, args.Next.Value, len(str))
		return gelo.BytesToSym(str[i : i+1])
	}
	idxs := make([]int, ac-1)
	build := make([]byte, ac-1)
	count := 0
	for i := args.Next; i != nil; i = i.Next {
		idxs[count] = ToIdx(vm, i.Value, len(str))
		count++
	}
	for i, v := range idxs {
		build[i] = str[v]
	}
	return gelo.BytesToSym(build)

}

func Serialize(_ *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.Null
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return w.Ser()
	})
}

func Count_substrings(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "count-substrings", "string substring", args)
	}
	s1 := args.Value.Ser().Bytes()
	s2 := args.Next.Value.Ser().Bytes()
	n, _ := gelo.NewNumberFromGo(bytes.Count(s1, s2))
	return n
}

var _split_parser = extensions.MakeOrElseArgParser("string ['on sep]?")

func Split(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	Args := _split_parser(vm, args)
	var sep = []byte(" ")
	if _, has := Args["on"]; has {
		sep = vm.API.LiteralOrElse(Args["sep"])
	}
	str := vm.API.LiteralOrElse(Args["string"])
	strs := bytes.Split(str, sep, -1)
	if len(strs) == 0 {
		return gelo.EmptyList
	}
	list := extensions.ListBuilder()
	for _, v := range strs {
		list.Push(gelo.BytesToSym(v))
	}
	return list.List()
}

var _join_parser = extensions.MakeOrElseArgParser("list ['with sep]?")

func Join(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	Args := _join_parser(vm, args)
	var sep = []byte("")
	if _, has := Args["with"]; has {
		sep = Args["sep"].Ser().Bytes()
	}
	list := vm.API.ListOrElse(Args["list"])
	slice := make([][]byte, list.Len())
	for count := 0; list != nil; list = list.Next {
		slice[count] = list.Value.Ser().Bytes()
		count++
	}
	return gelo.BytesToSym(bytes.Join(slice, sep))
}

func Starts_with(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "starts-with", "string prefix", args)
	}
	return gelo.ToBool(bytes.HasPrefix(
		args.Value.Ser().Bytes(),
		args.Next.Value.Ser().Bytes()))
}

func Ends_with(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "starts-with", "string prefix", args)
	}
	return gelo.ToBool(bytes.HasSuffix(
		args.Value.Ser().Bytes(),
		args.Next.Value.Ser().Bytes()))
}

func ToRunes(_ *gelo.VM, args *gelo.List, _ uint) gelo.Word {
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		is := w.Ser().Runes()
		if len(is) == 0 {
			return gelo.EmptyList
		}
		list := extensions.ListBuilder()
		var n *gelo.Number
		for _, v := range is {
			n, _ = gelo.NewNumberFromGo(v)
			list.Push(n)
		}
		return list.List()
	})
}

func FromRunes(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "->runes", "list of runes", args)
	}
	var buf bytes.Buffer
	bucket := make([]byte, 4)
	for rs := vm.API.ListOrElse(args.Value); rs != nil; rs = rs.Next {
		i, ok := vm.API.NumberOrElse(rs.Value).Int()
		if !ok || i < 0 || i > math.MaxInt32 {
			gelo.TypeMismatch(vm, "rune", "number")
		}
		length := utf8.EncodeRune(bucket, int(i))
		buf.Write(bucket[:length])
	}
	return gelo.BytesToSym(buf.Bytes())
}

func ToLower(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return gelo.BytesToSym(bytes.ToLower(w.Ser().Bytes()))
	})
}

func ToUpper(_ *gelo.VM, args *gelo.List, _ uint) gelo.Word {
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return gelo.BytesToSym(bytes.ToUpper(w.Ser().Bytes()))
	})
}

func _stripper_gen(f func([]byte, func(int) bool) []byte) func(gelo.Word) gelo.Word {
	return func(w gelo.Word) gelo.Word {
		return gelo.BytesToSym(f(w.Ser().Bytes(), unicode.IsSpace))
	}
}

var _left_stripper = _stripper_gen(bytes.TrimLeftFunc)
var _right_stripper = _stripper_gen(bytes.TrimRightFunc)
var _both_stripper = _stripper_gen(bytes.TrimFunc)
var _strip_parser = extensions.MakeOrElseArgParser("['left|'right]? string+")

func Strip(vm *gelo.VM, args *gelo.List, _ uint) gelo.Word {
	Args := _strip_parser(vm, args)
	strings := Args["string"].(*gelo.List)
	var mapper func(w gelo.Word) gelo.Word
	if _, ok := Args["left"]; ok {
		mapper = _left_stripper
	} else if _, ok := Args["right"]; ok {
		mapper = _right_stripper
	} else {
		mapper = _both_stripper
	}
	return strings.MapOrApply(mapper)
}

func Length(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		n, _ := gelo.NewNumberFromGo(len(w.Ser().Bytes()))
		return n
	})
}

func StrToList(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.EmptyList
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		bs := w.Ser().Bytes()
		if len(bs) == 0 {
			return gelo.Null
		}
		list := extensions.ListBuilder()
		for i, _ := range bs {
			list.Push(gelo.BytesToSym(bs[i : i+1]))
		}
		return list.List()
	})
}

func Nullp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.True
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return gelo.ToBool(gelo.IsNullString(w))
	})
}

//returns true if the string(s) only whitespace
func Emptyp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.True
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		lit := vm.API.LiteralOrElse(w)
		return gelo.ToBool(len(lit) == gelo.SlurpWS(lit, 0))
	})
}

var StringCommands = map[string]interface{}{
	"chars":            Chars,
	"<-string":         Serialize,
	"count-substrings": Count_substrings,
	"split":            Split,
	"join":             Join,
	"starts-with":      Starts_with,
	"ends-with":        Ends_with,
	"<-runes":          ToRunes,
	"->runes":          FromRunes,
	"<-lower":          ToLower,
	"<-upper":          ToUpper,
	"str->list":        StrToList,
	"strip":            Strip,
	"length":           Length,
	"null?":            Nullp,
	"empty?":           Emptyp,
}
