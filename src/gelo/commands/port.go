package commands

import "gelo"

func ChanCon(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 0 {
        gelo.ArgumentError(vm, "Chan", "", args)
    }
    return gelo.NewChan()
}

func PortClosex(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 1 {
        gelo.ArgumentError(vm, "close!", "port", args)
    }
    p := vm.API.PortOrElse(args.Value)
    if p.Closed() {
        gelo.RuntimeError(vm, "port already closed")
    }
    p.Close()
    return gelo.Null
}

func PortClosedp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 1 {
        gelo.ArgumentError(vm, "closed?", "port", args)
    }
    p := vm.API.PortOrElse(args.Value)
    return gelo.ToBool(p.Closed())
}

func PortWritex(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac < 2 {
        gelo.ArgumentError(vm, "write!", "port rest+", args)
    }
    p := vm.API.PortOrElse(args.Value)
    if p.Closed() {
        gelo.RuntimeError(vm, "attempted to write to a closed port")
    }
    var msg gelo.Word
    if ac == 2 {
        msg = args.Next.Value
    } else {
        msg = args.Next
    }
    p.Send(msg)
    return msg
}

func PortReadx(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 1 {
        gelo.ArgumentError(vm, "read!", "port", args)
    }
    p := vm.API.PortOrElse(args.Value)
    if p.Closed() {
        gelo.RuntimeError(vm, "attempted to read from a closed port")
    }
    return p.Recv()
}

var PortCommands = map[string]interface{} {
    "Chan":     ChanCon,
    "close!":   PortClosex,
    "closed?":  PortClosedp,
    "read!":    PortReadx,
    "write!":   PortWritex,
}
