For information on how to enable or disable traces, see GeloTracing.

Traces display a program in a more exact notation than plaintext before and after rewriting.

Traces represent lists as `[e0]->[e1]->...->[eN]->0`.

Since lines are lists both before and after rewriting, lines are rendered as above. Since clauses are lines in lines, they are displayed as such; that is `a [b c] d` is `[a]->[[b]->[c]->0]->[d]->0`. There is no ambiguity between this and a line with a list as an element as a line can only contain a list after rewriting and can only contain a clause before rewriting.

If a line contains a sigil, that element in the list is represented as `[$<[i]->0>]->`, similiarly for `@`. The fact that the element inside the sigil identifier `<`...`>` is a list is an artifact of parsing--it cannot actually contain multiple elements.

Note that `a "b c" d` traces as `[a]->[b c]->[d]->0` as "b c" is a single word.

A parsed quotation (that is, one being used as code) traces as
```
(v)->line0
(v)->line1
...
(v)->lineN
(0)
```

After rewriting, a line is rendered as a list with each element (the command and all arguments) [serialized](Ser.md). It is traced before the command is dereferenced in the namespace so `puts hello world` trace to `[puts]->[hello]->[world]->0` and not `[*ALIEN*]->[hello]->[world]->0`.

See RewritingMetaphor.