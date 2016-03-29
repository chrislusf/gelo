To write a Gelo command in Go, you need only write a function of the type `func(vm gelo.VM, args *gelo.List, ac uint) gelo.Word` and [UsingTheVM##Adding\_items\_to\_a\_VM:_The\_Register__**family register] it with any VMs you create, or put it in a bundle with your other [Alien](GeloTypes#Alien.md) commands. While there are few methods on `*gelo.VM` that can be called from within a [running VM](UsingTheVM.md), the vm object has two very useful objects: [Ns](Ns.md) and [API](API.md). With them and the APIs of the GeloTypes, you have everything you need to write a custom command.**

Let's examine the process of writing a command in detail by iteratively designing the `sgn` command defined in `gelo/commands/math.go`._

First, let's recall the definition of the `sgn` function. `sgn` is a real function that computes the sign of its argument: 0 for 0, 1 for positive numbers, and -1 for negative numbers.

Let's give it a go:
```
func Sgn(_ *gelo.VM, args *gelo.List, _ uint) gelo.Word {
  N := args.Value.(*gelo.Number) //args.Value is the first element in args
  n := N.Real() //unboxes N into a float64
  switch {
  case n < 0:
    n = -1
  case n == 0:
    n = 0
  case n > 0:
    n = 1
  }
  return gelo.NewNumber(n) //boxes n in a *gelo.Number
}
```
This works, as long as it gets at least one argument and that argument is a number. If it's not a number, Go will panic and the Gelo VM will not catch it as it is not a [Gelo error](GeloErrors.md), so this command could bring down not only the VM but possibly the program it is embedded in. Not good. The third parameter, `ac` is the length of `args`. And we can use an OrElse function to make sure that if `args.Value` isn't a `*gelo.Number` that a proper error is raised, which the VM can process safely. Let's give it another shot:
```
func Sgn(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
  if ac != 1 {
    //formats a runtime error showing the expected arguments vs what was recieved
    gelo.ArgumentError(vm, "sgn", "number", args)
  }
  //Since we can't do anything if not given a number, we panic
  n := vm.API.NumberOrElse(args.Value).Real()
  switch {
  case n < 0:
    n = -1
  case n == 0:
    n = 0
  case n > 0:
    n = 1
  }
  return gelo.NewNumber(n)
}
```

Now, even if misused this command acts responsibly and is a good citizen, but there's still a possible improvement.

It's common among GeloCommands that rewrite a single value to return a list of values if called with multiple (valid) arguments. The `*gelo.List` type provides a convience method `(*gelo.List).MapOrApply(f func(gelo.Word) gelo.Word) *gelo.List` which if the list is of length one it returns the result of f, for longer lengths it returns a list whose nth element is f applied to the nth element of the original list. For a list of length 0, the result is undefined as often something special must be done in the case of no arguments, so no sensible default makes sense. In our case, take the `sgn` of nothing makes no sense so we'll raise an error as before.
```
func Sgn(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
  if ac == 0 {
    gelo.ArgumentError(vm, "sgn", "number+", "")
  }
  return args.MapOrApply(func(w gelo.Word) gelo.Word {
    n := vm.API.NumberOrElse(w).Real()
    switch {
    case n < 0:
      n = -1
    case n == 0:
      n = 0
    case n > 0:
      n = 1
    }
  })
}
```
Which is exactly as the command is defined in `gelo/commands/math.go`

It is convention for a command that has no sensible value to return, to return the empty string "" which in [Alien](GeloTypes#Alien.md) commands may be denoted by `gelo.Null`.