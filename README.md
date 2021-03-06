
Gelo -- Go Extension Language

Installation instruction: go get code.google.com/p/gelo go install code.google.com/p/gelo/geli go install code.google.com/p/gelo/gelrun cp $GOPATH/src/code.google.com/p/gelo/examples/prelude.gel $GOPATH/bin

An extensible extension programmable programming language written in Go, for Go.

The syntax is simple and minimal, as is the execution model.

Gelo can be thought of as a very low level language for a very high level machine with an extensible instruction set (via WritingAlienCommands). Gelo provides a few simple rules for rewriting and executing an expression.

The language's initial state has zero commands and can run no programs. Adding built in commands is simple and writing new commands to integrate Gelo into your application is easy. Since there are no commands built into the language, and commands must be bound to a namespace, to be used, and the names are stored in UTF-8, the language can be fully localized with the rest of your application.

This control allows Gelo to be used as a simple configuration language, an executable specification for your web apps routing table, an easy-to-use macro language for that text editor you're writing, a DSL to describe your app's workflow, a full-fledged language for writing plugins and extensions, or even a complete programming language in its own right.

Gelo has no dependencies beyond the Go standard library; and even those are kept to a minimum.

The syntax of Gelo is very similiar to that of Tcl but execution is more like a Lisp/Scheme language. The RewritingMetaphor for execution lets you build a domain-specific language by layering small commands of increasing abstraction, similiar to a Forth system, and obviates many of the situations where a macro facility would be used.

There are some examples of actual Gelo code (which use GeloCommands and the prelude) in src/examples. src/gelrun has the code for the file intepreter and src/geli has the code for REPL. Note: you must copy src/examples/prelude.gel into $GOPATH/bin for geli and gelrun to function without the -no-prelude switch.

The wiki page on UsingTheVM is a handy reference when reading the code.

For writing Gelo commands in Go, the entire standard library in src/commands is a reference.
