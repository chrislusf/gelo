package commands

import (
    "gelo"
    "regexp"
)

type Regexp struct {
    regexp *regexp.Regexp
    source  gelo.Symbol
}

func (_ *Regexp) Type() gelo.Symbol {
    return gelo.StrToSym("*REGULAR-EXPRESSION*")
}

func (r *Regexp) Ser() gelo.Symbol {
    return r.source
}

func (r *Regexp) Copy() gelo.Word {
    return r
}

func (r *Regexp) DeepCopy() gelo.Word {
    return r
}

func (r *Regexp) Equals(w gelo.Word) bool {
    r2, ok := w.(*Regexp)
    if !ok {
        return false
    }
    return r2.source.Equals(r.source) //XXX not perfect but close enough
}

func ReOrElse(vm *gelo.VM, w gelo.Word) *Regexp {
    r, ok := w.(*Regexp)
    if !ok {
        gelo.TypeMismatch(vm, "regexp", w.Type())
    }
    return r
}

func ReCon(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 1 {
        gelo.ArgumentError(vm, "Re", "specification", args)
    }
    spec := vm.API.LiteralOrElse(args.Value)
    rex, err := regexp.Compile(string(spec))
    if err != nil {
        gelo.SyntaxError(vm, err.String())
    }
    return &Regexp{rex, gelo.BytesToSym(spec)}
}

func Re_matchp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 2 {
        gelo.ArgumentError(vm, "match?", "regexp string", args)
    }
    r := ReOrElse(vm, args.Value)
    s := args.Next.Value.Ser().Bytes()
    return gelo.ToBool(r.regexp.Match(s))
}

func Re_matches(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 2 {
        gelo.ArgumentError(vm, "matches", "regexp string", args)
    }
    r := ReOrElse(vm, args.Value)
    s := args.Next.Value.Ser().Bytes()
    matches := r.regexp.MatchSlices(s)
    head := &gelo.List{gelo.BytesToSym(matches[0]), nil}
    tail := head
    for _, v := range matches[1:] {
        tail.Next = &gelo.List{gelo.BytesToSym(v), nil}
        tail = tail.Next
    }
    return head
}

func Re_replace(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 3 {
        gelo.ArgumentError(vm, "replace-all", "regexp string replace", args)
    }
    r := ReOrElse(vm, args.Value)
    src := args.Next.Value.Ser().Bytes()
    repl := args.Next.Next.Value.Ser().Bytes()
    return gelo.BytesToSym(r.regexp.ReplaceAll(src, repl))
}

func Re_replace_by(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
    if ac != 3 {
        gelo.ArgumentError(vm, "replace-all", "regexp string replace-cmd", args)
    }
    r := ReOrElse(vm, args.Value)
    src := args.Next.Value.Ser().Bytes()
    repl := args.Next.Next.Value
    return gelo.BytesToSym(r.regexp.ReplaceAllFunc(src, func(s []byte) []byte {
        args := gelo.NewList(gelo.BytesToSym(s))
        return vm.API.InvokeCmdOrElse(repl, args).Ser().Bytes()
    }))
}

var RegexpCommands = map[string]interface{} {
    "Re":               ReCon,
    "re-match?":        Re_matchp,
    "re-matches":       Re_matches,
    "re-replace":       Re_replace,
    "re-replace-by":    Re_replace_by,
    "re?":              Rep,
}
