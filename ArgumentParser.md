# Introduction #

The ArgumentParser is essentially a primitive regular expression matcher. Instead of matching over strings it matches over [lists](GeloTypes#List.md). It is far from complete. It uses a greedy longest sequence matcher so the equivalent to the regular expression `A.*B` cannot be matched. However, it is already extrememly useful. It allows you to create remarkably expressive commands. Before we continue to examine the ArgumentParser in greater detail, an example may better demonstrate its utility. The `if` command defined in GeloCommands uses the ArgumentParser to parse the traditional `if` syntax using this specification string:
```
cond 'then cons ['elif cond 'then cons]* ['else alt]?
```
(Note that the if part of the if "statement" is encoded by the command name itself)

# The basic idea #

The parser reads each item in the list according to its specification. It stores the result of parsing in a `map[sting]gelo.Word` if the [list](GeloTypes#List.md) can be parsed according to the specification. If the argument is too long, too short, or does not meet the specification, it returns an error.

# Specification strings #

The ArgumentParser is given a specification string that tells it how to parse arguments. An empty specification string only matches a [list](GeloTypes#List.md) of zero length.

The most basic term is a single word, called a variable matcher. In this case, the parser simply stores any item in the [list](GeloTypes#List.md) in the results or returns an error if there is nothing left in the [list](GeloTypes#List.md). For example, the specification `"one two three"` parses only lists of length three storing the first item in `one`, the second in `two`, and the third in `three`. Note that if the same word is used more than once as in `"A B A"` that `A` will be a list of the first and last elements in the [list](GeloTypes#List.md) to parse.

The second most basic is a single word prefixed by the literal prefix operator `'` like `'litereal`. This matches only items whose [serialization](Ser.md) is, byte-for-byte, the specified literal. For example, the specification `"one 'two three"` parses only lists of length three whose middle element [serializes](Ser.md) to `two`. If the [list](GeloTypes#List.md) is parsed the value of `two` in the returned `map` is `gelo.Null`.

Next are the postfix operators `*`, `+`, and `?`. They map directly to the familiar regular expression operators. As noted above, they greedily match the longest possible sequence. Examples:
  * `"one two three*"` will match [lists](GeloTypes#List.md) of length two or more. If the [list](GeloTypes#List.md) is of length two `three` will not be set in the returned `map`. Otherwise, it will be set to a [list](GeloTypes#List.md) of the remaining items in the [list](GeloTypes#List.md).
  * `"one two three+"` is the same as `"one two three three*"`.
  * `"one two three?"` matches [lists](GeloTypes#List.md) of length two or three. If the length is 2, `three` will be unset, and if the length is 3, `three` will contain the last element.

The alternation (left-associative binary) operator `|` chooses the first pattern that matches. The specification `"'a|'b|'c|'d"` matches a [list](GeloTypes#List.md) of length one only if it contains an item that literally matches `a`, `b`, `c`, or `d`.

Finally, the grouping bracket `[]` matches sub-specifications atomically. The specification `"one ['two three]? four"` matches [lists](GeloTypes#List.md) of length 2 or 4. If the length is 2, it can contain any two elements, but, if the length is 4, then the second element must match the literal `two`.

# An Extended Example #
Now that we've covered the specifications all quite familiar operators, let's go back to the `if` parser and analyze it in detail.
```
cond 'then cons ['elif cond 'then cons]* ['else alt]?
```

Now that we know how to read an ArgumentParser specification string it should make some semblance of sense, but it never hurts to study simple things in detail.

The first three elements are boring. It's the last two groups whose interactions are subtle and interesting. Clearly both are optional so that
```
#recall that quotes are one word
if {= 1 1} then { puts hello world }
#this results in the following map:
# cond => {= 1 1}
# cons => { puts hello world }
#or
if $var then 1 ;# returns "" if var derefs to false, incidentially
```
are representives of [lists](GeloTypes#List.md) that meet the minimum requirements given by the specification.

Now let's consider the last group. Clearly this allows for
```
if {= $x 1} then {
    something
} else {
    or-other
}
```
which sets `else` to `gelo.Null` and `alt` to the quote containing `or-other` in the returned `map`.

Now the question remains as to how the first group ineracts with the rest. `*` matches greedily so `"a b* c"` will fail--`b` eats all of the list and leaves nothing for `c` to match. But we see that `"a 'b* c"` will match `A b b b b b b C` since `*` fails when its operand `'b` fails leaving `C` to be matched. So `['a b]* 'c` will match `a 1 a 2 a 3 c` happily and stop matching. Thus we see that since the greedy longest sequence matching isn't a problem in cases where we can start groups with a literal.

Putting this in the context of our example, consider
```
if {cond1} then {
    A
} elif {cond2} then {
    B
} elif {cond3} then 5 else 6
```
If we parse `{cond1} then {\n\tA\n}`, the values of the `map` at that time are

|cond|`{cond1}`|
|:---|:--------|
|then|`""`     |
|cons|`{\n\tA\n}`|

Next we try the group `['elif cond 'then cons]*`.  The `*` parser tries the `[]` parser which tries its subparsers in sequence. We match `elif` and store it. Now we match a variable but that variable is already in the `map` so we take it out, and replace it with a list of the removed value followed by the current value. Next we match `then`. Since it's a literal we do not promote it to a list and leave it as is. Now the map is:

|cond|`[{cond1}]->[{cond2}]->0`|
|:---|:------------------------|
|then|`""`                     |
|cons|`[{\n\tA\n}]->[{\n\tB\n}]->0`|
|elif|`""`                     |

Similiarly, we match `elif {cond3} then 5` resulting in:

|cond|`[{cond1}]->[{cond2}]->[{cond3}]->0`|
|:---|:-----------------------------------|
|then|`""`                                |
|cons|`[{\n\tA\n}]->[{\n\tB\n}]->[5]->0`  |
|elif|`""`                                |

Now, the next token is `else` so the first group fails when `else` fails to match the literal `elif` so we move to the `?` parser whose group matches and we end up with:

|cond|`[{cond1}]->[{cond2}]->[{cond3}]->0`|
|:---|:-----------------------------------|
|then|`""`                                |
|cons|`[{\n\tA\n}]->[{\n\tB\n}]->[5]->0`  |
|elif|`""`                                |
|else|`""`                                |
|alt |`6`                                 |

# Using the ArgumentParser #
There are two ways to make an ArgumentParser from a specification string:
  * `MakeArgParser(spec string) func(*gelo.List) (map[string]gelo.Word, bool)`
> > Takes a string and returns a function that takes a [list](GeloTypes#List.md) and parses it  according to the specification and returns a `map` of the results if the second return is `true` or `nil` if the second return is `false`.
  * `MakeOrElseArgParser(spec string) func(*gelo.VM, *gelo.List) map[string]gelo.Word`
> > This wraps the above and returns a function that takes a `*gelo.VM` so that it can panic via a [ArgumentError](GeloErrors.md) on failure.