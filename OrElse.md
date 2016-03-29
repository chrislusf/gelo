The `*OrElse*` metafamily is a Gelo Go-coding convention to create functions that panic with an appropriate [error](GeloErrros.md) to reduce boiler plate for common situations, such as "Name not in Namespace" => `(*VM).Ns.LookupOrElse(Symbol) Word` or "argument must be a `*Number`" => `(*VM).API.NumberOrElse(Word) *Number`.

As an example for unbelievers, we write an [Alien](GeloTypes#ALIEN.md) command that takes one argument that must be a symbol, looks it up in the current namespace, ensures it is a number, and returns that number. First we write it without any `*OrElse*` functions and then, again, with them.

Without:
```
import . "gelo"

func Verify_numeric(vm *VM, args *List, ac uint) Word {
  if ac != 1 {
    ArgumentError(vm, "verify-numeric", "name-of-a-number", args)
  }

  sym, ok := args.Value.(Symbol)
  if !ok {
    TypeMismatch(vm, "symbol", args.Value.Type())
  }

  deref, there := vm.Ns.Lookup(sym)
  if !there {
    VariableUndefined(vm, sym)
  }

  num, ok := deref.(*Number)
  if !ok {
    //Can we coerce it to a number?
    num, ok = NewNumberFrom(deref)
    if !ok {
      TypeMismatch(vm, "number", deref.Type()
    }
  }

  return num
}
```

While that doesn't look so bad, this is a trivial command and you will be writing that code over and over regardless of the complexity of the command. True, it is clearer, but the errors must all be handled in exactly the same way with exactly the same code. The `*OrElse*` metafamily replaces this boilerplate with a clean, declarative approach to common error conditions that is just as readable. Moreover, it could be easy to forget something like the `NewNumberFrom` check resulting in some commands being able to, for example, handle a number coming directly from stdin and some saying that `7` is a `Symbol`.

Now let's look at the same command as above with its equivalent `*OrElse*` code.

```
import . "gelo"

func Verify_numeric(vm *VM, args *List, ac uint) Word {
  if ac != 1 {
    ArgumentError(vm, "verify-numeric", "name-of-a-number", args)
  }

  sym   := vm.API.SymbolOrElse(args.Value)
  deref := vm.Ns.LookupOrElse(sym)
  return vm.API.NumberOrElse(deref)
}
```

Or if you're feeling really perverse:

```
import . "gelo"

func Verify_numeric(vm *VM, args *List, ac uint) Word {
  if ac != 1 {
    ArgumentError(vm, "verify-numeric", "name-of-a-number", args)
  }

  return vm.API.NumberOrElse(vm.Ns.LookupOrElse(vm.API.SymbolOrElse(args.Value)))
}
```

Which I of course cannot approve of no matter how much code like that you may find in my code base.