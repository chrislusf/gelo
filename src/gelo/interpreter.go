package gelo

func (vm *VM) _deref(item *sNode) (ret Word) {
	switch item.tag {
	case synClause:
		cmd_ret := vm._eval_line(item.val.(*sNode))
		switch t := cmd_ret.(type) {
		default:
			TypeMismatch(vm, "symbol or quote", cmd_ret.Type())
		case Quote:
			ret = t.Ser()
		case Symbol:
			ret = t
		}
	case synQuote:
		ret = item.val.(Quote).Ser()
	case synLiteral:
		ret = item.val.(Word)
	default:
		systemError(vm, "invalid node type dereferenced--parser incorrect",
			item)
	}
	ret = vm.Ns.LookupOrElse(ret)
	run_trace("derefed", item, "=>", ret)
	return
}

func (vm *VM) rewrite(c *sNode) (*List, uint) {
	var head, tail *List
	ac := uint(0)
	fill := func(item Word) {
		if head != nil {
			ac++
			tail.Next = &List{item, nil}
			tail = tail.Next
		} else {
			//initial word is the command, so we don't inc ac
			head = &List{item, nil}
			tail = head
		}
	}
	for cmd := c; cmd != nil; cmd = cmd.next {
		switch cmd.tag {
		case synLiteral, synQuote:
			fill(cmd.val.(Word))
		case synIndirect:
			fill(vm._deref(cmd.val.(*sNode)))
		case synClause:
			fill(vm._eval_line(cmd.val.(*sNode)))
		case synSplice:
			c := cmd.val.(*sNode)
			var splice Word
			if c.tag == synClause {
				//cannot use dereference because we expect the result to be
				//a list not a symbol
				splice = vm._eval_line(c.val.(*sNode))
			} else {
				splice = vm._deref(c)
			}
			s, ok := splice.(*List)
			if !ok {
				RuntimeError(vm, "Attempted to splice nonlist")
			}
			for ; s != nil; s = s.Next {
				fill(s.Value)
			}
		}
	}
	run_trace("rewrote", c, "to", head)
	return head, ac
}

func (vm *VM) peval(line *List, ac uint) (ret Word, c *command, args *List) {
	ret, args = line.Value, line.Next
	if q, ok := ret.(Quote); ok {
		//if the head is a quote we optimistically mark it invokable
		ret = q.unprotect()
	} else if _, ok = ret.(Alien); !ok {
		//Not a quote or alien, we attempt to dereference the serialization
		//of the command and had better get a quote or alien (or defer)
		run_trace("evaluating named command", ret)
		switch cmd := vm.Ns.LookupOrElse(ret).(type) {
		default:
			TypeMismatch(vm, "invokable", ret.Type())
		case Quote:
			//dereferenced a quote so we assume that the value is invokable
			ret = cmd.unprotect()
		case Alien:
			ret = cmd
		case *defert:
			// it is up to the caller to register the defer or report an
			// error if a defer cannot be used here
			ret = BI_defer
		}
	} else if d, ok := ret.(*defert); ok {
		//very rare special case
		ret = d
		return
	}

	//either an anonymous alien (like the result of something like the compose
	// command in gelo/commands/combinators.go)
	if gocmd, ok := ret.(Alien); ok {
		run_trace("invoking alien")
		//it is up to gocmd to mark the quote invokable if it wishes
		ret = gocmd(vm, args, ac)
		args = nil
	}

	//either the head or the result of one of the above
	if q, ok := ret.(*quote); ok {
		run_trace("invoking quote")
		c, ok = q.fcode()
		ret = nil
		if !ok {
			//attempted to invoke a literal, nonempty quote,
			//we reparse and let any syntax errors bubble up to the top
			//so the user can see what went wrong and where
			panic(force_synerr(vm, q))
		}
	}
	return
}

func (vm *VM) _eval_line(line *sNode) Word {
	w, c, args := vm.peval(vm.rewrite(line))
	//if we are only evaluating a single line that means we're in a clause
	if _, ok := w.(*defert); ok {
		RuntimeError(vm, "defer commands must not be in a clause")
	}
	if c != nil {
		return vm.eval(c, args)
	}
	return w
}

var argument_sym = interns("arguments")

func (vm *VM) eval(script *command, arguments *List) (ret Word) {
	if script == nil {
		//handle Noop
		return Null
	}
	var kill bool
	var c *command
	for script != nil {
		//store arguments
		//TODO use something else so we can tail recurse
		//without blowing the defer stack
		ns := vm.cns
		oldargs, there := ns.get(argument_sym)
		ns.set(argument_sym, arguments)
		if there {
			defer ns.set(argument_sym, oldargs)
		} else {
			defer ns.del(argument_sym)
		}
		run_trace("evaluation started")
		if script.next != nil { //allows 1 liners
			for ; script.next != nil; script = script.next {
				ret, c, arguments = vm.peval(vm.rewrite(script.cmd))
				if _, ok := ret.(*defert); ok {
					//attach a defer handler
					if arguments == nil {
						ArgumentError(vm, "defer", "No command to defer", "")
					}
					defer func(arguments *List) {
						run_trace("invoking defer:", arguments)
						ac := uint(arguments.Len())
						_, c, arguments = vm.peval(arguments, ac)
						if c != nil {
							vm.eval(c, arguments)
						}
						run_trace("defer handler invoked")
					}(arguments)
					run_trace("attached defer handler:", arguments)
				} else if c != nil {
					//not a defer, but got code
					vm.eval(c, arguments)
				}
			}
		}
		//tail call (or 1 liner)
		ret, script, arguments = vm.peval(vm.rewrite(script.cmd))
		if _, ok := ret.(*defert); ok {
			RuntimeError(vm, "defer call cannot be in tail position")
		}
		//before we continue, see if anyone wants to kill us
		if _, kill = <-vm.kill_switch; kill {
			panic(kill_control_code(byte(0)))
		}
	}
	run_trace("evaluation became", ret)
	return ret
}
