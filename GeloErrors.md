# Kinds of Errors #
## internal errors ##
There are only two errors inside the Gelo language. Runtime errors and syntax errors. Unlike most languages, it is quite possible for syntax errors to be issued at runtime (see CodeDataDuality.) Internally, these erorrs are panic'd but they are recovered by the VM or, inside a script, via one of the BuiltIns. A number of predefined errors are defined [below](GeloErrors#Predefined_errors.md).

## external errors ##
There are also three kinds of errors that exist outside the Gelo language: they all represent bugs and will cross the gelo package boundary, for reasons to be explained below.

The first is `_errSystem`. These are raised inside the VM when internal invariants are broken and hence they represent a serious bug in the VM itself. Hopefully, no one will ever see one unless they're working on the VM.

Second, is the cruelly named `_hostProgrammerError`. These are fired when a programmer embedding a VM in an application violates a contract with the API, such as  calling [Convert](GeloConversions#Generic_conversion.md) with a type it does not understand or attempting to perform an unsafe operation on a [dead VM](UsingTheVM.md).

Finally, there are panics that emanate from an [alien](GeloTypes#Alien.md) command that are not of any of the above types. This is a bug in that command and rather than hide it the VM intentionally allows them to escape to be handled by the embedder.

Generally, it is bad form to allow a panic to cross package boundaries, but the three errors above that, so to speak, leak through the package boundary all do so for a very good reason: they are bugs.

`_errSystem` is a bug in the VM itself and it should poke its head up as soon as possible so that it may be rightly struck down and to make sure that you, the embedder, are not made responsible for our errors by releasing a surely fine program crippled by our incompetence.

A `_hostProgrammerError` would likely happen anyway but by firing a custom error we can provide a detailed explanation allowing you to more quickly repair the minor fault in your  embedding of the Gelo VM and be surer that it is correct.

Finally, all other panics are allowed to escape because silently eating them would allow a  bug in an [alien](GeloTypes#Alien.md) command to hide, increasing the likelihood that your scripts are incorrect--we can't have that.

# Predefined Errors #

Errors cannot be created explicitly, they are created by these functions:
  * `SyntaxError(info ...interface{})`
  * `RuntimeError(vm *gelo.VM, info ...interface{})`

These functions do not return. They create and populate their respective error type and then panic it. I wish I could say that there was some grand design decision for this, maybe something about declartive programming, but, honestly, I just got tired of typing `panic(SyntaxError("blah blah"))`

Yes, you can create your own syntax errors. Gelo is very DSL friendly.

These convience methods wrap RuntimeError and format error messages for common classes of errors. Many of them are called implicitly by OrElse methods on failure. GeloCommands defines more wrappers around `RuntimeError`.
  * `VariableUndefined(vm *gelo.VM, name interface{})`
  * `TypeMismatch(vm *gelo.VM, expected, recieved interface{})`
  * `ArgumentError(vm *gelo.VM, command_name, arg_spec, recieved interface{})`