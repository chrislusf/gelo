# Word #

Every value in Gelo is a Word, however no value is just a Word. It has the following methods:

|`Type() Symbol`|The name of the type of this value|
|:--------------|:---------------------------------|
|`Ser() Symbol` |The [serialization](Ser.md) of this value|
|`Copy() Word`  |A copy of the value, shallow in the case of collections, or the value itself if the type is immutable|
|`DeepCopy() Word`|For a collection, return a copy of the collection with `DeepCopy` called on all of its elements. Otherwise, the same as `Copy`|
|`Equals(Word) bool`| See if paramter is equal to value, based on the type's idea of equality|

Feel free to define your own.

# Bool #

A boolean value. Equality is defined if and only if both comparators are `Bool` and have the same truth value.

Defines the aliases `True` and `False`.

# Symbol #

A mutable byte string. Equality is defined by `bytes.Equals` against the [serialization](Ser.md) of the other word.

Defines an alias for the string "", `Null`.

# Number #

A `*Number` is an immutable float64. Equality is defined by `float64 == float64`.

# Quote #

An immutable block of code or an immutable string. Equality is defined if and only if both comparators are `Quote` and `bytes.Equals` returns true given the source of each `Quote`.

Defines an alias for the no-operation, `{}`, Noop.

# Port #

A `Port` is an object with the following operations:
  * Send(Word)
  * Recv() Word
  * Close()
  * Closed() bool

GeloCore defines on kind of `Port`, `*Chan`.

## Chan ##

A `*Chan` is a wrapper around a `chan Word`. Equality is defined by if both comparators are the same object.

# [List](List.md) #

A [\*List](List.md) is a collection backed by a singly-linked list. This is the type of `arguments`. A `Word` is equal to a list if it is also a list and is of the same length and all e<sub>i</sub> in the first list are equal to all o<sub>i</sub> in the second list.

Declares an alias `EmptyList` for `(*List)nil`.

See also: ListBuilder

# [Dict](Dict.md) #

A [\*Dict](Dict.md) is a collection backed by a Go `map`. Keys are [serialized](Ser.md) and converted to string so that they may be hashed. Values are stored as is.

A `Word` is equal to a `*Dict` if it is also a `*Dict` of the same length and for every key-value pair contained therein, the key exists in both dictionaries and both values are `Equal`.

# Alien #

A command written in native Go. It is simply a function with the signature: `func(*VM, *List, uint) Word`. Aliens are equal by `==` in Go. See WritingAlienCommands.

# Error #

An `Error` is either a runtime error (`ErrRuntime`) or a syntax error (`ErrSyntax`). Normally, an error halts the VM and [returns the error from the VM](UsingTheVM#Executing_Programs.md). But Gelo code run through [eval, safe-eval, or go](BuiltIns.md) will be rewritten to an `Error` if there is an error. See RewritingMetaphor, GeloErrors.

Errors are equal if they are the same kind of error with the same message.


---

See also: GeloConversions