package commands

import (
	"bytes"
	"code.google.com/p/gelo"
)

func Type_of(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "type-of", "value+", "")
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return w.Type()
	})
}

func _make_tpred(type_sig string) gelo.Alien {
	sig := []byte(type_sig)
	return func(_ *gelo.VM, args *gelo.List, ac uint) gelo.Word {
		if ac == 0 {
			return gelo.False
		}
		return args.MapOrApply(func(w gelo.Word) gelo.Word {
			return gelo.ToBool(bytes.Equal(sig, w.Type().Bytes()))
		})
	}
}

var Listp, Dictp = _make_tpred("*LIST*"), _make_tpred("*DICT*")
var Symbolp, Portp = _make_tpred("*SYMBOL*"), _make_tpred("*PORT*")
var Quotep, Boolp = _make_tpred("*QUOTE*"), _make_tpred("*BOOL*")
var Alienp, Nump = _make_tpred("*ALIEN*"), _make_tpred("*NUMBER*")
var Syntax_errorp = _make_tpred("*SYNTAX-ERROR*")
var Runtime_errorp = _make_tpred("*RUNTIME-ERROR*")

//defined here since we have this lovely machine, but used in the bundle defined
//in regexp.go
var Rep = _make_tpred("*REGULAR-EXPRESSION*")

var TypePredicates = map[string]interface{}{
	"type-of":        Type_of,
	"list?":          Listp,
	"dict?":          Dictp,
	"symbol?":        Symbolp,
	"port?":          Portp,
	"quote?":         Quotep,
	"bool?":          Boolp,
	"number?":        Nump,
	"alien?":         Alienp,
	"syntax-error?":  Syntax_errorp,
	"runtime-error?": Runtime_errorp,
}
