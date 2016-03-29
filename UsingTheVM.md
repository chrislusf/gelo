# Anatomy of a VM #

Every VM has the following:
  * A locally unique ID (that is, unique in a single run of the host application)
  * An I/O [Port](GeloTypes#Port.md), akin to stdin and stdout on a process
  * A program
  * An [API](API.md) object
  * A namespace and associated [Ns](Ns.md) object

The [API](API.md) and [Ns](Ns.md) objects have methods for [Alien](GeloTypes#Alien.md) commands to access the VM. They should never be called from outside an [Alien](GeloTypes#Alien.md) and  [Alien](GeloTypes#Alien.md) methods will deadlock if they call the methods declared directly on `*VM`, except for the following:
  * `(*VM).Spawn() *VM`
  * `(*VM).ProcID() uint32`
  * `(*VM).IsDead() bool`
  * `(*VM).IsRunning() bool`
  * `(*VM).IsIdle() bool`
Though calling the last three is fairly useless as the only possibility is that the VM is running, in that case.

# Creating and destroying virtual machines #

There are two ways to create a virtual machine. The first is to call `NewVM(io Port) *VM`.

The second is to spawn a child VM from an existing VM by calling `(*VM).Spawn() *VM`. A spawned VM inherits its parents I/O [Port](GeloTypes#Port.md) and has read -- but not write -- access to its parent's namespace.

A VM created by `NewVM` is said to be a **top-level** VM, and a VM created by calling another VM's `Spawn` is said to be a **spawned** VM. A VM that is not spawned is said to be a **top level** VM.

A VM that is not executing a program is said to be **idle**. A VM that has been killed is a **dead** VM and any attempt to use it will result in a panic.  A VM that is neither dead nor idle is said to be **running**.

To check if a VM is dead, call `(*VM).IsDead() bool`. To check if a VM is running, call `(*VM).IsRunning() bool`. To check if a VM is idle, call `(*VM).IsIdle() bool`.

There are two ways to make a VM dead. The first is to call its `(*VM).Destroy()` method. The second is to call `Kill(*VM)`.

`Destroy` must be called from the goroutine that owns the VM. `Kill` must not be called from the goroutine that owns the VM.


# Adding items to a VM: The Register`*` family #

A freshly created top-level VM has no items in its namespace. To add items to a VM's namespace externally, you call one of
  * `(*VM).Register(string, interface{})`
  * `(*VM).RegisterBundle(map[string]interface{})`
  * `(*VM).RegisterBundles([]map[string]interface{})`

In all three, the `interface{}` values must be either a Gelo [type](GeloTypes.md) or a Go type that is [convertible](GeloConversions.md) to a Gelo [type](GeloTypes.md). Internally, they call [Convert](GeloConversions#Generic_conversion.md) so they can panic if given an unknown type.

Registering an item with a running VM will block and not complete until the VM halts. Attempts to add a new value to a dead VM panics.

# Getting items from a VM: The Read`*` family #

To get the value of an item in a VM's namespace, use a member of the Read`*` family. All members of the family take a string that is the name to look up and return a copy of the requested value and a bool that is false if either the value is not found or the requested conversion could not be performed. Attempting to read from a running VM will block until the VM halts.
|`(*VM).ReadWord(string) (Word, bool)`|No conversions.|
|:------------------------------------|:--------------|
|`(*VM).ReadString(string) (string, bool)`|Returns the item's [serialization](Ser.md) as a string.|
|`(*VM).ReadRunes(string) ([]int, bool)`|Returns the item's [serialization](Ser.md) as runes.|
|`(*VM).ReadBytes(string) ([]byte, bool)`|Returns the item's [serialization](Ser.md) as a byte slice.|
|`(*VM).ReadBool(string) (bool, bool)`|Attempts to convert a [Bool](GeloTypes#Bool.md) to a `bool`.|
|`(*VM).ReadMap(string) (map[string]Word, bool)`|Attempts to convert a [\*Dict](GeloTypes#Dict.md) to a `map[string]Word`.|
|`(*VM).ReadSlice(string) ([]Word, bool)`|Attempts to convert a [\*List](GeloTypes#List.md) to a `[]Word`.|
|`(*VM).ReadQuote(string) (Quote, bool)`|Returns a [Quote](GeloTypes#Quote.md).|
|`(*VM).ReadPort(string) (Port, bool)`|Returns a [Port](GeloTypes#Port.md).|
|`(*VM).ReadChan(string) (*Chan, bool)`|Returns a [\*Chan](GeloTypes#Chan.md).|
|`(*VM).ReadInt(string) (int64, bool)`|Attempts to convert to a [\*Number](GeloTypes#Number.md). Returns false if the [\*Number](GeloTypes#Number.md) cannot be represented as an `int64`.|
|`(*VM).ReadFloat(string) (float64, bool)`|Attempt to convert to [\*Number](GeloTypes#Number.md) and return its `float64` value.|

# Executing Programs #

All methods below must only be called on an idle VM. If they are called on a dead VM, they will panic.

The simplest way for an idle VM to execute code is `(*VM).Do(string) (Word, Error)`, where  the `string` is meant to be a fixed ideal string. If the string contains a syntax error, that panic is not caught by `Do` as `Do` is intented only for **fixed** programs and as such will **not** trap syntax errors. Invoking `Do` does not alter the VM's program.

A VM needs a program to execute. You can set it with `(*VM).SetProgram(Quote) Error`. The error is non-nil if the [Quote](GeloTypes#Quote.md) does not contain valid Gelo code (that is, it is a literal quote).

If you do not have a [Quote](GeloTypes#Quote.md), you can set the program from an `io.Reader` by calling `(*VM).ParseProgram(io.Reader) Error`, where the error is non-nil if what is `Read` from the `io.Reader` does not parse. The symmetric method is `(*VM).GetProgram() Quote`, which returns `nil` if no program has been set yet. `GetProgram` may safely be called on a dead VM.

If a program has been set, it can be executed by `(*VM).Exec(interface{}) (Word, Error)`. If no program is set `Exec` panics. The `interface{}` parameter is the `arguments` and must either be a Gelo [type](GeloTypes.md) or a Go type that is [convertible](GeloConversions.md) to a Gelo [type](GeloTypes.md). If there are no errors `Exec` returns the value of the last line in its program in its first return value. Otherwise, it retuns the error.

Finally, `(*VM).Run(io.Reader, interface{}) (Word, Error)` is a convience method that is equivalent to calling `ParserProgram` followed by `Exec`. It does not have its own locking so it is possible for a lock to sneak in between its calls to `ParseProgram` and `Exec`, so it should be avoided in a multithreaded environment.

If a VM is killed by a separate process while it is running, the execution will halt when the signal is recieved and `Exec`, `Run`, and `Do` return an Error whose `String` method returns "VM killed".

# Methods on `*VM` #
  * `(*VM).Spawn() *VM` (1)
  * `(*VM).Destroy()`
  * `(*VM).IsDead() bool` (1) (2)
  * `(*VM).IsRunning() bool` (1) (2)
  * `(*VM).IsIdle() bool` (1) (2)
  * `(*VM).Redirect(Port) Port`
> > Change the I/O `Port` of a VM. Only call when idle and not dead.
  * `(*VM).ProcID() uint32` (1) (2)
  * `(*VM).Register(string, interface{})`
  * `(*VM).RegisterBundle(map[string]interface{})`
  * `(*VM).RegisterBundles([]map[string]interface{})`
  * `(*VM).ReadWord(string) (Word, bool)`
  * `(*VM).ReadString(string) (string, bool)`
  * `(*VM).ReadBytes(string) ([]byte, bool)`
  * `(*VM).ReadRunes(string) ([]int, bool)`
  * `(*VM).ReadBool(string) (bool, bool)`
  * `(*VM).ReadMap(string) (map[string]Word, bool)`
  * `(*VM).ReadSlice(string) ([]Word, bool)`
  * `(*VM).ReadQuote(string) (Quote, bool)`
  * `(*VM).ReadPort(string) (Port, bool)`
  * `(*VM).ReadChan(string) (*Chan, bool)`
  * `(*VM).ReadInt(string) (int64, bool)`
  * `(*VM).ReadFloat(string) (float64, bool)`
  * `(*VM).SetProgram(Quote) Error`
  * `(*VM).GetProgram() Quote`
  * `(*VM).ParseProgram(io.Reader) Error`
  * `(*VM).Do(string) (Word, Error)`
  * `(*VM).Exec(interface{}) (Word, Error)`
  * `(*VM).Run(io.Reader, interface{}) (Word, Error)`

1 - Safe to be called on a running VM.
2 - Safe to be called on a dead VM.