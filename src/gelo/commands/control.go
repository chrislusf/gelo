package commands

import "gelo"

func _condition(vm *gelo.VM, w gelo.Word) bool {
	return vm.API.BoolOrElse(vm.API.InvokeWordOrReturn(w)).True()
}

var _if_parser = MakeOrElseArgParser(
	"cond 'then cons ['elif cond 'then cons]* ['else alt]?")

func If(vm *gelo.VM, args *gelo.List, ac uint) (ret gelo.Word) {
	Args := _if_parser(vm, args)
	cond, cons := Args["cond"], Args["cons"]
	if lcond, ok := cond.(*gelo.List); !ok { //no elifs
		if _condition(vm, cond) {
			return vm.API.TailInvokeWordOrReturn(cons)
		}
	} else {
		for lcons := cons.(*gelo.List); lcond != nil; lcond, lcons = lcond.Next, lcons.Next {
			if _condition(vm, lcond.Value) {
				return vm.API.TailInvokeWordOrReturn(lcons.Value)
			}
		}
	}
	if _, ok := Args["else"]; ok {
		return vm.API.TailInvokeWordOrReturn(Args["alt"])
	}
	return gelo.Null
}

/*
 * case-of val ['as var] ['by command] {
 *      pattern1 => result1
 *      pattern2 => result2
 *      ...
 *      patternN => resultN
 *      [otherwise resultN+1]
 * }
 *
 * If you specify "as var" it binds the result to var in any patterns and
 * results that are commands, otherwise val will only be bound to arguments as
 * per usual.
 * If "by command" is specified, patterns are matched against the result of
 * [command val] as arguments (not var) and any results that are commands
 * recieve val as arguments (and var).
 * Check val against patterns 1..N returning the respective result. If the Nth
 * pattern is a sequence like "a b c" then the items are matched left to right
 * and if any of them match it resultN is returned. If there are no matches
 * return resultN+1 if there is an otherwise clause and "" if there isn't.
 * Pattern items are matched by their Equals method.
 */
func _case_eval(vm *gelo.VM, w gelo.Word, args *gelo.List) (ret gelo.Word) {
	//XXX This disallows us to make tail calls
	switch w.(type) {
	case gelo.Quote, gelo.Alien, gelo.Symbol:
		ret, err := vm.API.InvokeCmd(w, args)
		if err != nil {
			return w
		}
		return ret
	}
	return w
}
func _cases_synerr() {
	gelo.SyntaxError("Patterns needs to be:",
		"\"value+ => resultant\", where value may be a command")
}

var _cases_parser = MakeOrElseArgParser("value ['as var]? ['by cmd]? cases")

func Case_of(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	Args := _cases_parser(vm, args)

	key := Args["value"]
	arguments := gelo.AsList(key)
	cases, ok := vm.API.PartialEval(vm.API.QuoteOrElse(Args["cases"]))

	if !ok || cases.Next == nil && cases.Value == nil {
		gelo.SyntaxError(vm, "Expected:", _cases_parser,
			"{[value+ => resultant\n]+ [otherwise result]?} Got:", args)
	}

	//Give val a name, specified by var, in clauses of the cases block
	//XXX This disallows us from making tail calls
	if var_name, there := Args["var"]; there {
		name := vm.API.SymbolOrElse(var_name)
		if d, there := vm.Ns.DepthOf(name); there && d == 0 {
			defer vm.Ns.Set(name, vm.Ns.LookupOrElse(name))
		} else {
			defer vm.Ns.Del(name)
		}
		vm.Ns.Set(name, key)
	}

	//run val through cmd before comparing
	if cmd, there := Args["cmd"]; there {
		key = vm.API.InvokeCmdOrElse(cmd, arguments)
	}

	//Parse lines
	for ; cases != nil; cases = cases.Next {

		//if last line, see if it's the otherwise clause
		if cases.Next == nil {
			item := cases.Value.(*gelo.List)
			s, ok := item.Value.(gelo.Symbol)
			if ok && gelo.StrEqualsSym("otherwise", s) {
				if item.Next == nil || item.Next.Next != nil {
					_cases_synerr()
				}
				return _case_eval(vm, item.Next.Value, arguments)
			}
		}

		item := cases.Value.(*gelo.List)

		//line is too short
		if item.Next == nil || item.Next.Next == nil {
			_cases_synerr()
		}

		//Parse a single line
		var head, tail *gelo.List
		var resultant gelo.Word
		for ; item != nil; item = item.Next {
			if item.Next == nil { //ultimate cell
				resultant = item.Value
			} else if item.Next.Next == nil { //penultimate cell
				s, ok := item.Value.(gelo.Symbol)
				if ok && gelo.StrEqualsSym("=>", s) {
					continue
				}
				_cases_synerr()
			} else if head != nil { //regular cell, store
				tail.Next = &gelo.List{item.Value, nil}
				tail = tail.Next
			} else {
				head = &gelo.List{item.Value, nil}
				tail = head
			}
		}

		//see if key matches any of the items we found on this line
		for ; head != nil; head = head.Next {
			if key.Equals(head.Value) {
				return _case_eval(vm, resultant, arguments)
			}
		}
	}

	return gelo.Null //no match, no otherwise
}

var ControlCommands = map[string]interface{}{
	"if":      If,
	"case-of": Case_of,
}
