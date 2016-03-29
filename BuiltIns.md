GeloCore only defines four commands. All others may be implemented using API, see WritingAlienCommands. Builtin commands' Go names are prefixed by `BI_`. The `gelo` package also defines a bundle, `Core`:
```
var Core = map[string]interface{}{
	"eval":      BI_eval,
	"safe-eval": BI_safe_eval,
	"go":        BI_go,
	"defer":     BI_defer,
}
```

We will use the bundle's names for these commands in the discussion below.

`eval code args*` and `safe-eval code args*` evaluate their `code` as usual, passing it `args` as its `arguments`. If `code` dies with a [runtime or syntax error](GeloErrors.md) or is halted, instead of halting the entire VM, `eval` or `safe-eval` is rewritten to that error. `eval` evaluates its code in the current VM, meaning that it has free access to modify the Namespace. `safe-eval` spawns a child VM to evaluate the code in, thus denying write access to the current Namespace. Neither will catch a panic that is not a Gelo Runtime or Syntax error, to avoid hiding latent bugs in an Alien command.

`go [--redirect port]? code args*` is essentially the same as `safe-eval` except the VM is run in its own goroutine. The `--redirect port` argument allows you to set the VM's IO Port.

`defer line` executes line after the enclosing command is [rewritten](RewritingMetaphor.md) ("returns"). `line` will be rewritten where it appears in the program so:
```
set! n 4
defer puts $n
unset! n
```
will result in a 4 being written to the VM's IO Port, whereas:
```
set! n 4
defer { puts $n }
unset! n
```
will cause a variable undefined runtime error to be raised. Unlike in Go, it is not possible to change the return value with a defered invocation; however, combined with [tail calls](TailCall.md), `defer` becomes quite powerful. Consider writing a command like a Lisp `with-` macro:
```
command with-file {fname 'as name 'do block} {
  set! $name [open-file $fname rw]
  defer {
    close! $name
    unset! name
  }
  block
}
```
Which can be used like:
```
with-file example.txt as file do {
  write! $file "hello, world!"
}
```
Which is equivalent to:
```
set! file [open-file example.txt rw]
write! $file "hello, world"
close! $file
unset! file
```
Of course, there should be error handling to check that name wasn't unset and isn't already closed and it would be helpful if I had actually implemented the commands for dealing with files.