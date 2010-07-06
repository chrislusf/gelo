package commands

import "gelo"
import "bytes"

func Puts(vm *gelo.VM, args *gelo.List, _ uint) gelo.Word {
    if args == nil {
        return gelo.Null
    }
    var buf bytes.Buffer
    var ret gelo.Word
    for ; args.Next != nil; args = args.Next {
        buf.Write(args.Value.Ser().Bytes())
        buf.WriteString(" ")
    }
    buf.Write(args.Value.Ser().Bytes())
    ret = gelo.BytesToSym(buf.Bytes())
    vm.API.Send(ret)
    return ret
}

func Gets(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 0 {
        gelo.ArgumentError(vm, "gets", "", args)
    }
    return vm.API.Recv()
}

var IOCommands = map[string]interface{} {
    "puts": Puts,
    "gets": Gets,
}
