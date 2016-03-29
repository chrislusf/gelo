Execution of a Gelo program can be thought of as succesively rewriting a program until it is a single value. Reading the below it may seem to be terribly inefficient and repetitive. Realize that this metaphor is how to conceptualize the execution of a program and the actual  implementation is allowed to perform all kinds of dirty low down tricks to keep things chugging along smooth, as long as execution appears to follow this metaphor for all intents and purposes. We will explain the basics and follow that with examples to help digest those ideas.

# The basics #

### Quotes ###
A Gelo program is a [quote](GeloTypes#Quote.md). (Note that while the outermost `{` and `}` are omitted from the text of a program they are there conceptually.) A quote is rewritten by first rewriting all the escapes in the quote and then rewriting each line, top to bottom. The quote itself is rewritten to the value of rewriting its last line.

### Lines ###
A line is a sequence of words terminated by a `;` or a newline. If the line is a comment, everything until the next newline is ignored and rewriting continues on the next line. Rewriting a line is a two stage process. First, each word is rewritten, left to right, until it is in **normal form**; that is, no more words can be rewritten and are all basic types. Second, the first word of the sequence is dereferenced in the [namespace](Ns.md), unless it's a quote, which is itself invokable. If the dereferenced value is non-invokable or the quote fails to be rewritten, the program halts and is rewritten to an error. Otherwise, the invokable is invoked with the rest of the words as its arguments, and the line is rewritten to result of that invocation. A line may be written on more than one physical line in a number of ways: by using the ignore whitespace escape `\*`, by having a multiline string literal such as
```
a very "long
line" that spans two physical lines
```
or a quote such as
```
a line that {
  spans multiple lines
}
```

### Words ###
A word is rewritten dependent on its syntatic type:
  * If a word is literal, a symbol, quote, or number, it is written as is.
  * If a word is prefixed by `$`, the rest of the word is rewritten and then the whole word is rewritten to the value of the normal form of the rest of the world in the namespace.
  * If a word is prefixed by `@`, the rest of the word is rewritten and the value of dereferencing the value of the normal form of the rest of word in place. If the dereferenced value is not a list the program is rewritten to an error. Splicing writes the values of the list in place.
  * If a word is clause, that is an embedded line bracketed by `[` and `]`, the word is rewritten to the result of rewriting the embedded line.
There are a few caveats to the above. Namely a word prefixed by a sigil (`$` or `@`) it may be followed by a clause (`[]`) but no further sigils are parsed. That is, `$$a` looks up the value of `$a` in the namespace instead of performing a double indirection. A double indirection could be faked by `$[id $a]`, assuming the GeloCommands `id` is in the namespace. No rewriting is done inside `""`, so `"$a"` is rewritten to the symbol `$a` and not the value of `a` in the namespace; like wise "a b" is rewritten to one element and not two.


# Examples #

We give the examples by writing the input program in plaintext as you would and TraceFormat to make it clear how it is parsed; then we show the various stages of rewriting in plaintext followed by TraceFormat so that you can see how it works.

We separate the plaintext from the TraceFormat by writing `<plain>` :: `<trace>`. We write snippets of traces such as `[value]->` to denote that the element is not part of a full list (which would be terminated by 0).

### Words ###

The interplay of these rewrites won't be clear until the next section, but there is value in seeing them in isolation.

If `a` is a literal, then the word `a` :: `[a]->` is rewritten to `a` :: `[a]->`.

If the value of `a` in the namespace is `hello`, then the word `$a` :: `[$<[a]->0>]->` is rewritten to `hello` :: `[hello]->`.

If the value of `a` in the namespace is the list `{hello, world}`, then the word `$a` :: `[$<[a]->0>]->` is rewritten to `{hello, world}` :: `[[hello,]->[world]->0]->`.

If the value of `a` in the namespace is the list `{hello, world}`, then the word `@a` :: `[@<[a]->0>]->` is rewritten to `hello, world` :: `[hello,]->[world]->`.

If the value of `a` in the namespace is a command that always returns the number 4, then the word `[a]` :: `[[a]->0]->` is rewritten to `4` :: `[4]->`.

If the value of `a` in the namespace is a command that always returns `b` and `b` in the namespace is the number 4, then the word `$[a]` :: `[$<[[a]->0]->0>]->` is rewritten to `$b` :: `[$<[b]->0>]->` is rewritten to `4` :: `[4]->`.

### Lines ###

Consider the line `$cmd [id 4] @rest you see.` :: `[$<[cmd]->0>]->[[id]->[4]->0]->[@<[rest]->0>]->[you]->[see.]->0` where `cmd` in the namespace is `puts`, `id` is the identity function, and `rest` is the list `{is a number}` :: `[is]->[a]->[number]->0`.

First, we rewrite the words in the line, left to right. `$cmd` is rewritten to `puts`, yielding: `puts [id 4] @rest you see.` :: `[puts]->[[id]->[4]->0]->[@<[rest]->0>]->[you]->[see.]->0`. Then, `[id 4]` is rewritten to `4`, yielding: `puts 4 @rest you see.` :: `[puts]->[4]->[@<[rest]->0>]->[you]->[see.]->0`. `@rest` is rewritten to `is a number` yielding: `puts 4 is a number you see.` :: `[puts]->[4]->[is]->[a]->[number]->[you]->[see.]->0`. Finally, we invoke the `puts` command with the arguments `4 is a number you see.` and the line is rewritten to the result of that invocation, which in the case of `puts` is to simply return the list after outputting it to the VMs IO [Port](GeloTypes#Port.md).

To clarify the difference between `$` and `@` Let's see what the above line would be rewritten to if we substituted `$rest` for `@rest`: `puts 4 {is a number} you see.` :: `[puts]->[4]->[[is]->[a]->[number]->0]->[you]->[see.]->0`.

### Quotes ###

A quote is rewritten line by line top to bottom and becomes the value of the result of the last line. So:
```
{
  $cmd [id 4] @rest you see.
}
```
will behave exactly as it did in the first example in the section above. However, that is not the full story. Quote are only rewritten when you ask for them to be rewritten, as we implicitly did above by having it be the first (and only) element of the line. Furthermore, quotes can represent code or data so if we change the last example to
```
puts {
  $cmd [id 4] @rest you see.
}
```
then the quote is not treated as code but data and the output will be:
```

  $cmd [id 4] @rest you see.

```
(note that the whitespace is preserved and that the outer `{}` were elided).

In fact, when treated as data not even escaping is performed. The only requirement for a quote used as data is that all `{` and `}` be balanced (it will not count `\{` or `\}` toward this though it does not escape) so that the parser knows where to stop. Incidentially, this is why `{` and `}` must be balanced in comments. Also note that using a quote as data does not preclude it being used as code or vice versa, though a data quote that does not contain a valid program will result in an error if evaluated, of course.

Because quotes are not evaluated unless told to you may pass code to a command and let it choose to evaluate it as it sees fit.