package commands

import (
	"gelo"
	"gelo/extensions"
)

func DictCon(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.NewDict()
	}
	if ac != 1 {
		gelo.ArgumentError(vm, "Dict", "{{k1 v1} {k2 v2} ... {kn vn}}", args)
	}
	d, ok := gelo.UnserializeDictFrom(args.Value)
	if !ok {
		gelo.RuntimeError(vm, "Cannot unserialize", args.Value, "into a dict")
	}
	return d
}

func Dict_get(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if !(ac == 2 || ac == 3) {
		gelo.ArgumentError(vm, "dict.get", "dictionary key default?", args)
	}
	d := vm.API.DictOrElse(args.Value)
	k := args.Next.Value
	ret, ok := d.Get(k)
	if !ok {
		if ac == 3 {
			//return default value
			return args.Next.Next.Value
		}
		gelo.RuntimeError(vm, "dictionary does not contain", k)
	}
	return ret
}

// Like get but requires that the default is specified and if key is not found
// sets key to the default in addition to returning it
func Dict_getx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 3 {
		gelo.ArgumentError(vm, "dict.get!", "dictionary key default", args)
	}
	d := vm.API.DictOrElse(args.Value)
	k := args.Next.Value
	ret, ok := d.Get(k)
	if !ok {
		df := args.Next.Next.Value
		d.Set(k, df)
		return df
	}
	return ret
}

func Dict_setx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 3 {
		gelo.ArgumentError(vm, "dict.set", "dictionary key value", args)
	}
	d := vm.API.DictOrElse(args.Value)
	k := args.Next.Value
	v := args.Next.Next.Value
	d.Set(k, v)
	return v
}

func Dict_unsetx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "dict.unset!", "dictionary key", args)
	}
	d := vm.API.DictOrElse(args.Value)
	k := args.Next.Value
	v, ok := d.Get(k)
	if !ok {
		gelo.RuntimeError(vm,
			"dict.unset! attempted to unset a key not in dictionary")
	}
	d.Del(k)
	return v
}

func Dict_setp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "dict.setp?", "dictionary key", args)
	}
	d := vm.API.DictOrElse(args.Value.(*gelo.Dict))
	k := args.Next.Value
	return gelo.ToBool(d.Has(k))
}

func Dict_keys(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "dict.keys", "dictionary", args)
	}
	d, list := vm.API.DictOrElse(args.Value), extensions.ListBuilder()
	for k, _ := range d.Map() {
		list.Push(gelo.StrToSym(k))
	}
	return list.List()
}

func Dict_values(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "dict.values", "dictionary", args)
	}
	d, list := vm.API.DictOrElse(args.Value), extensions.ListBuilder()
	for _, v := range d.Map() {
		list.Push(v)
	}
	return list.List()
}

func Dict_items(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "dict.values", "dictionary", args)
	}
	d, list := vm.API.DictOrElse(args.Value), extensions.ListBuilder()
	for k, v := range d.Map() {
		list.Push(gelo.NewList(gelo.NewList(gelo.StrToSym(k), v)))
	}
	return list.List()
}

//Add two dictionaries. d1 = d1 + d2 where
// For k, v in d2, d1[k] = d2[k] if k not in d1 (d2[k] is not copied)
// d1 + d2 /= d2 + d1
func Dict_add(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "dict.add", "dictionary1 dictionary2", args)
	}
	d1 := vm.API.DictOrElse(args.Value)
	d2 := vm.API.DictOrElse(args.Next.Value)
	for k, v := range d2.Map() {
		if !d1.StrHas(k) {
			d1.StrSet(k, v)
		}
	}
	return d1
}

//Substract two dictionaries, d1 = d1 - d2 where
// For k, _ in d2, remove d1[k]
// d1 - d2 /= d2 - d1
func Dict_sub(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "dict.sub", "dictionary1 dictionary2", args)
	}
	d1 := vm.API.DictOrElse(args.Value)
	d2 := vm.API.DictOrElse(args.Next.Value)
	for k, _ := range d2.Map() {
		if d1.StrHas(k) {
			d1.StrDel(k)
		}
	}
	return d1
}

var _dict_commands = map[string]gelo.Alien{
	"set!":   gelo.Alien(Dict_setx),
	"set?":   gelo.Alien(Dict_setp),
	"unset!": gelo.Alien(Dict_unsetx),
	"get":    gelo.Alien(Dict_get),
	"get!":   gelo.Alien(Dict_getx),
	"add":    gelo.Alien(Dict_add),
	"sub":    gelo.Alien(Dict_sub),
	"keys":   gelo.Alien(Dict_keys),
	"values": gelo.Alien(Dict_values),
	"items":  gelo.Alien(Dict_items),
}

func DictAggregate(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac < 2 {
		gelo.ArgumentError(vm, "dict", "dictionary command args*", args)
	}
	d := args.Value
	command := args.Next.Value.Ser().String()
	rest := args.Next.Next
	if _, ok := _dict_commands[command]; !ok {
		gelo.ArgumentError(vm, "dict",
			"dictionary 'set!|'unset!|'get|'get!|'set?|'keys|'value args*",
			command)
	}
	return _dict_commands[command](vm, &gelo.List{d, rest}, ac-1)
}

func DictToCommand(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "dict->command", "dictionary", args)
	}
	m := vm.API.DictOrElse(args.Value).Map()
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		if ac == 0 {
			gelo.ArgumentError(vm, "command generated by dict->command",
				"argument+", args)
		}
		name := args.Value.Ser().String()
		if v, ok := m[name]; ok {
			if args.Next != nil {
				return vm.API.TailInvokeCmd(v, args.Next)
			} else {
				return vm.API.TailInvokeWordOrReturn(v)
			}
		} else if args.Next != nil {
			gelo.RuntimeError(vm, name, "is not a valid subcommand")
		} else {
			gelo.RuntimeError(vm, name, "is not a valid subcommand or entry")
		}
		panic("command generated by dict->command in impossible state") //Issue 65
	})
}

func Zip_map(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "zip-map", "key-list value-list", args)
	}
	keys := vm.API.ListOrElse(args.Value)
	values := vm.API.ListOrElse(args.Next.Value)
	m := make(map[string]gelo.Word)
	for ; keys != nil || values != nil; keys, values = keys.Next, values.Next {
		m[keys.Value.Ser().String()] = values.Value
	}
	return gelo.NewDictFrom(m)
}

var DictCommands = map[string]interface{}{
	"Dict":          DictCon,
	"dict":          DictAggregate,
	"dict->command": DictToCommand,
	"zip-map":       Zip_map,
}
