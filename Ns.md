See NamespaceSpecification for more information about how namespaces work on a conceptual level.

The most basic description, for those in a hurry, is that a namespace is a pair consisting of a pointer to a parent namespace and a [Dict](GeloTypes#Dict.md) of values at its level.

In general, "the namespace" refers to the current level and all of its parents, while "the current namespace" refers to only the dictionary at this level and not any of its parents.

If a VM has a parent, accessing a value from the parent's namespace deep copies it to the topmost namespace of the child VM. If a child VM attempts, attempts to delete an entry from a namespace owned by a parent, that name is blacklisted in the child VM and can never be accessed by that child or, by extension, any of its children again, but the value remains unharmed to the parent.

### Sidebar: levels ###
Some `Ns` methods accept an integer parameter to specify the level of the operation.
  * If the level is 0, the operation is perfomed on the current namespace.
  * If the level is postive, the operation will be attempted `n` levels up. If there are less than `n` levels the operation will fail. If the operation is a mutator, and there are les than `n` level owned by the VM the operation will fail.
  * If levels is less than 0, the operation will be perfomed on the topmost namespace that the VM owns for mutators or the topmost namespace that exists for queries and accessors.

# Queries #
  * `(*VM).Ns.Lookup(key Word) (Word, bool)`
> > Search namespace for instance of `key` and found return its value, if present.
  * `(*VM).Ns.Get(level int, key Word) (Word, bool)`
> > Search only the namespace at `level` and none of its parents or children for `key`.
  * `(*VM).Ns.Locals(levels int) *Dict`
> > Create a `Dict` from the key value pairs in each namespace implied by `levels`. If `levels` is not 0, `Locals` starts at the current namespace and, as it proceeds up the namespace chain, it will not add any key that has already been added; that is, shadowing is preserved under this operation. Values owned by the VM will not be copied, and, as usual, values that are now owned will be deep copied.

# Mutators #
  * `(*VM).Ns.Set(levels int, key, value Word) bool`
> > Insert a key-value pair into the namespace specified by `levels`, fails if `levels` specified a namespace that does not exist or is not owned by the requesting VM.
  * `(*VM).Ns.Inject(levels int, dict *Dict) bool`
> > Insert all key-value pairs in `dict` into the namespace at `levels`. Fails if `levels` specifies a namespace that does not exist or that the requestor does not own. It is the callers responsibility to ensure that `dict` will not be written to by other goroutines. `Inject` does not copy values from the `dict`.
  * `(*VM).Ns.Del(key Word) (Word, bool)`
> > Search the namespace for `key`. If found delete it and return its value. Returns false if there is no entry for `key` at any level.
  * `(*VM).Ns.Mutate(key, value Word) bool`
> > Look up name and replace its value, returns false if name is not found in namespace and  in this case does **not** set anything.
  * `(*VM).Ns.MutateBy(Word, func(Word)(Word,bool)) (Word, bool)`
> > Similiar to `Mutate` except instead of blindly replacing the value it uses the specified transformation function if and only if the transformation returns true. `MutateBy` returns true if the name was found, mutated or not. Because of the, locking mechanisms internal to the [namespace](NamespaceSpecification.md), it is absolutely imperative that **no** namespaces are touched in any way inside the transformation function.
  * `(*VM).Ns.Swap(n1, n2 Word) (w2, w1 Word, bool)`
> > Lookup and swap `n1` and `n2`. If both are found, and say `n1=w1` and `n2=w2`, return `(w2, w1, true)`. If only one is found or neither are found, return `(Null, Null, false)`.

# Creation and destruction #
  * `(*VM).Ns.Fork(*namespace)`
> > Change current namespace. Given `nil` creates a fresh namespace. The level 1 namespace will be the current namespace previous to the call.
  * `(*VM).Ns.Unfork() (*namespace, bool)`
> > If this is the topmost namespace of VM, returns `(nil, false)` and leaves the current namespace untouched. Return the current namespace and set the current namespace to its parent.

# `*`Depth`*` family #
All of these methods start counting at 0 (the current namespace).
  * `(*VM).Ns.Depth() int`
> > Return the number of namespaces between here and the topmost namespace.
  * `(*VM).Ns.LocalDepth() int`
> > Like `Depth` but stops counting once we reach the parent's ns. If the VM has no parent, this is exactly `Depth`.
  * `(*VM).Ns.DepthOf(Symbol) (int, bool)`
> > Return the depth of name or false if name is not defined.