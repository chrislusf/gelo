package commands

import (
	"bytes"
	"gelo"
	"gelo/extensions"
)

func _args_to_spec(args *gelo.List, ac uint) string {
	var spec string
	switch ac {
	case 0:
		spec = ""
	case 1: // cases like "a b c*" or {'on|'off name}
		spec = args.Value.Ser().String()
	default: //input is a b c*
		var buf bytes.Buffer
		for ; args != nil; args = args.Next {
			buf.Write(args.Value.Ser().Bytes())
			buf.WriteString(" ")
		}
		spec = string(buf.Bytes())
	}
	return spec
}

func ArgumentParser(_ *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	parser := extensions.MakeOrElseArgParser(_args_to_spec(args, ac))
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, _ uint) gelo.Word {
		return gelo.NewDictFrom(parser(vm, args))
	})
}

func MaybeArgumentParser(_ *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	parser := extensions.MakeArgParser(_args_to_spec(args, ac))
	return gelo.Alien(func(vm *gelo.VM, args *gelo.List, _ uint) gelo.Word {
		Args, ok := parser(args)
		if !ok {
			return gelo.NewList(gelo.False)
		}
		return gelo.NewList(gelo.True, gelo.NewDictFrom(Args))
	})
}

var ArgParserCommands = map[string]interface{}{
	"ArgumentParser":      ArgumentParser,
	"MaybeArgumentParser": MaybeArgumentParser,
}
