Every VM has an API object that defines helper methods meant only to be called from [Alien](GeloTypes#Alien.md) commands. **Calling an API method outside of an [Alien](GeloTypes#Alien.md) being executed by the VM will result in undefined behavior**.


# Conversions, the `*`OrElse family #

`(*VM).API.`TYPE`OrElse(Word)` TYPE
> The following are all equivalent to testing type assertions and firing off a [TypeMismatch](GeloErrors.md) error if they fail. They save a great deal of boilerplate code. Any additional coercions are noted.

  * `(*VM).API.NumberOrElse(Word) *Number`
> > Additionally attempts to convert the arguments [serialization](Ser.md) to a `*Number` if simple type assertion fails.

  * `(*VM).API.QuoteOrElse(Word) Quote`
  * `(*VM).API.ListOrElse(Word) *List`
> > If type assertion fails attempts to [UnserializeListFrom](GeloConversions#ListConstruction.md).
  * `(*VM).API.DictOrElse(Word) *Dict`
> > If type assertion fails attempts to [UnserializeDictFrom](GeloConversions#DictConstruction.md)
  * `(*VM).API.PortOrElse(Word) Port`
  * `(*VM).API.ChanOrElse(Word) Port`
> > (It is not a typo that `ChanOrElse` returns `Port`: confer [the Port specification](GeloTypes#Port.md))
  * `(*VM).API.BoolOrElse(Word) Bool`
  * `(*VM).API.SymbolOrElse(Word) Symbol`
  * `(*VM).API.AlienOrElse(Word) Alien`
  * `(*VM).API.InvokableOrElse(Word) Word`
> > See `IsInvokable` in [Miscellany](API#Miscellany.md) below.
  * `(*VM).API.LiteralOrElse(Word) []byte`
> > Accepts either a `Quote` or `Symbol` and converts it to a `[]byte`. If you merely want a `[]byte` and do not care about the original type call `a_value.Ser.Bytes()` instead.

# The `*`Invoke`*` family #
No rewriting is performed by any `*Invoke*` family member, arguments must be in [normal form](GeloSpec#INVOCATION.md).

## The Invoke`*` subfamily ##

  * `(*VM).API.Invoke(*List) (Word, Error)`
> > The head of the list is the command to invoke and the tail is the arguments to invoke the command with.
  * `(*VM).API.InvokeCmd(Word, *List) (Word, Error)`
> > Same as above, with the command and arguments explicitly separated.
  * `(*VM).API.InvokeCmdOrElse(Word, *List) Word`
> > Same as above, but panic on error. See: OrElse
  * `(*VM).API.InvokeWordOrReturn(Word) (Word, Error)`
> > If the `Word` `IsInvokable`, invoke and return the result and error of invocation. If the `Word` is not invokable, simply return as-is.

## The TailInvoke`*` subfamily ##
Like `Invoke*` but allows tail calls to be optimized away, at the expense of not being able to capture errors. It is up to the caller to immediately return the value returned by these.
  * `(*VM).API.TailInvoke(*List) Word`
  * `(*VM).API.TailInvokeCmd(Word, *List) Word`
  * `(*VM).API.TailInvokeWordOrReturn(Word) Word`

# PartialEval #
`(*VM).API.PartialEval(Quote) (*List, bool)`

PartialEval rewrites but does not evaluate the lines of its argument. If it is given a quote that does not contain valid code, or evaluation results in an error, `PartialEval` returns false. Each element in the list is itself a list. The first item in the returned list is the first line in the quote after rewriting and so on.

Example:
`PartialEval` on the following quote (assuming all the variables and commands are defined)
```
    @[$very-indirect] atom $deref
    an ordinary line
    a [somewhat] ordinary line
```
results in (in TraceFormat)
```
[[result]->[of]->[very]->[indirect]->[atom]->[result-of-deref]->0]->[[an]->[ordinary]->[list]->0]->[[a]->[[result-of-somewhat ]->[ordinary]->line]->0]->0
```
and as plain text, for easier reading:
```
    result of very indirect atom result-of-deref
    an ordinary line
    a result-of-somewhat ordinary line
```

# Miscellany #
  * `(*VM).API.IsInvokable(w Word) (i Word, ok bool)`
> > If `w` is a `Symbol`, look up in namespace. If there's no such entry in the namespace, return `(nil, false)`, otherwise see if the dereferenced value is a code `Quote` or `Alien` and return that `Quote` or `Alien` and true if so. If `w` is not a `Symbol`, see if it's a `Quote` or `Alien` and if so return it and true. Otherwise, return `(nil, false)`.
  * `(*VM).API.Trace(...interface{})`
> > Display a [trace](GeloTracing.md), if **Alien\_trace**'s are enabled.
  * `(*VM).API.Halt(*List)`
> > Halt the VM, without error, returning its argument as the result of the program
  * `(*VM).API.Recv() Word` and `(*VM).API.Send(Word)`
> > Proxies for the VM's I/O [Port](GeloTypes#PORT.md).