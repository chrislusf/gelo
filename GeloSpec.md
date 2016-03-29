# LANGUAGE SPECIFICATION #


## LEXICAL STRUCTURE AND SYNTAX ##

### SPECIAL CHARACTERS ###
Any UTF-8 byte in the following, not including the period: `@$\[]{}";#`. They partition naturally into five classes.

  * One is [sigils](GeloSpec#SIGIL.md): `@$`.
  * The other is [brackets](GeloSpec#BRACKET.md): `[]{}"`.
  * The `;` is a singleton partition that is an alias for [line](GeloSpec#LINE.md) termination.
  * The next is the singleton `#`, used for [comments](GeloSpec#COMMENT.md).
  * The last is the singleton [escape](GeloSpec#ESCAPE.md): `\`.

There are situations in which some of the special characters are not special and can be used like a non-special character. They are detailed below, where appropriate.

### ESCAPE ###
The backslash character, `\`, followed by any character. The character following `\` is to be known as the escaped character. A [special character](GeloSpec#SPECIAL_CHARACTERS.md) is said to be taken literally if it is escaped.

An escape is a static substitution performed at parse time.
  * If the escaped character is one of the following letters, `abfnrtv`, it has the standard meaning.
  * If the escaped character is the asterisk, `*`, the parser erases all whitespace until the next nonwhitespace character.
  * Otherwise, the parser ignores any meaning that it would otherwise attribute to the escaped character.

It is not an error to escape a character without meaning. In the string [bracket](GeloSpec#BRACKET.md), `""`, the only escapes that are processed are `\*` and `\"` and any other backslash is taken literally. If a [quote](GeloSpec#QUOTE.md) bracket `{}` is taken literally, no escapes are processed but escaped `{}` do not count toward the requirement that all `{` must be matched with `}`.

### LINE ###
A sequence of [words](GeloSpec#WORD.md) terminated by newline, `\n`, or the semicolon, `;`. The length of a line is the number of words it contains.

### COMMENT ###
A comment is a special [line](GeloSpec#LINE.md) whose first nonwhitespace character is the pound symbol `#`. To use `#` as the first character of a [line](GeloSpec#LINE.md) literally it must be [escaped](GeloSpec#ESCAPE.md). If it is not the first character of a [line](GeloSpec#LINE.md), there is no need to escape it. There must be balanced `{}` in comments. This allows for multiline comments.

### WORD ###
A [bracket](GeloSpec#BRACKET.md) or UTF-8 character sequence delimited by whitespace or juxtapostion with a [bracket](GeloSpec#BRACKET.md). Any word may be prefixed by a [sigil](GeloSpec#SIGIL.md). Any [special character](GeloSpec#SPECIAL_CHARACTERS.md) to be literally included in a word must be escaped. Any UTF-8 character that is not a [special character](GeloSpec#SPECIAL_CHARACTERS.md) or whitespace may be in the character sequence comprising the word without escaping. A literal word is a word that does not have a [sigil](GeloSpec#SIGIL.md) and is not a [bracket](GeloSpec#BRACKET.md). If the word is not a [bracket](GeloSpec#BRACKET.md) or [sigil](GeloSpec#SIGIL.md), the length of a word is defined as the number of characters the word comprises, after escaping is performed.

### SIGIL ###
The **splice**, `@`, or the **substitution**, `$`, prefixing a [word](GeloSpec#WORD.md). The [word](GeloSpec#WORD.md) prefixed by a sigil is said to be the [word](GeloSpec#WORD.md) of that sigil. In a nonprefix position the sigil characters do not need to be escaped to be used as literals. The length of a sigil is defined as the length of the word it prefixes. A sigil must prefix a [word](GeloSpec#WORD.md).

### BRACKET ###
A sequence of characters between an opening bracket, one of `[{"` and a closing bracket, one of `"}]`. A sequence between the `[]` brackets is a [clause](GeloSpec#CLAUSE.md). A sequence between the `{}` brackets is a [quote](GeloSpec#QUOTE.md). A sequence between the `""` brackets is a word that is delimited by `"` instead of whitespace, including newlines; no [escapes](GeloSpec#ESCAPE.md) are processed unless the escaped character is `"` itself or `*`. The length of a `"` bracketed [word](GeloSpec#WORD.md) is the length of the characters between `""` afer escaping. The only word of length 0 is represented by the literal `""`. Unbalanced, unescaped brackets are a syntax error.

### CLAUSE ###
A clause is a [word](GeloSpec#WORD.md) that is a [line](GeloSpec#LINE.md) bracketed by `[]`. The length of a clause is the length of the enclosed [line](GeloSpec#LINE.md).

### QUOTE ###
A [word](GeloSpec#WORD.md) that is a sequence of zero or more [lines](GeloSpec#LINE.md) bracketed by `{}`. If a quote is used as a command, the usual escaping rules apply. If it is used as a literal [word](GeloSpec#WORD.md), no escaping is performed. Even if a quote is taken literally, all `{` must be matched by a properly nested `}`. Though [escapes](GeloSpec#ESCAPE.md) are not processed, escaped `{` or `}` do not count toward nesting. Though it need not be bracketed, the entirety of the input program is a quote. The length of a quote used literally is the length of the characters it contains, not including the outermost `{}`. A quote is a literal until it is used as code.


## EVALUATION ##

The [words](GeloSpec#WORD.md) in a [line](GeloSpec#LINE.md) are evaluated from left to right, and the [line](GeloSpec#LINE.md) is rewritten to the result of its evaluation.

The [lines](GeloSpec#LINE.md) of a [quote](GeloSpec#QUOTE.md) are evaluated from top to bottom, and the [quote](GeloSpec#QUOTE.md) is rewritten to the result of evaluating its last [line](GeloSpec#LINE.md).

### REWRITING ###

Before a [line](GeloSpec#LINE.md) is evaluated, it's [words](GeloSpec#WORD.md) are rewritten:
  * A literal [word](GeloSpec#WORD.md) or any [quote](GeloSpec#QUOTE.md) is written as-is.
  * A [substitution](GeloSpec#SIGIL.md) is rewritten as the result of dereferencing its [word](GeloSpec#WORD.md).
  * A [splice](GeloSpec#SIGIL.md) dereferences its [word.](GeloSpec#WORD.md) The result of dereferencing must be a list. Then, the list is expanded in place, increasing the number of arguments by one less the length of the list. If the list is empty, nothing is written, decreasing the number of arguments by one.
  * A [clause](GeloSpec#CLAUSE.md) is written as the result of executing its enclosed [line](GeloSpec#LINE.md). If either [sigil](GeloSpec#SIGIL.md) prefixes the [clause](GeloSpec#CLAUSE.md), it takes as input the result of rewriting the [clause](GeloSpec#CLAUSE.md). If a [clause](GeloSpec#CLAUSE.md) is prefixed by a [splice](GeloSpec#SIGIL.md), the result of the [clause](GeloSpec#CLAUSE.md) must be a list.

### INVOCATION ###

After rewriting a [line](GeloSpec#LINE.md) is said to be in **normal form** and the first [word](GeloSpec#WORD.md) in a [line](GeloSpec#LINE.md) is the **command** and the rest are the **arguments** of the command.
  * If the command is a [quote](GeloSpec#QUOTE.md) or an alien--a command written in Go code--it is invoked with those arguments.
  * Otherwise, the command is a literal [word](GeloSpec#WORD.md) that is dereferenced. The result of dereferencing must be either a [quote](GeloSpec#QUOTE.md) or alien. If it is not a [quote](GeloSpec#QUOTE.md) or alien, evaluation ends in an error.

If the command is a [quote](GeloSpec#QUOTE.md), dereferencing the word `arguments` in its body returns the arguments the callee invoked the [quote](GeloSpec#QUOTE.md) with.