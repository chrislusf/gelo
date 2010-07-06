===LANGUAGE SPECIFICATION===


=LEXICAL STRUCTURE AND SYNTAX=

SPECIAL CHARACTERS: Any UTF-8 byte in the following, not including the period:
    @$\[]{}". There are three partitions. One is sigils: @$. The other is
    brackets: []{}". The last is the singleton escape: \.

ESCAPE: The backslash character, \, followed by any character. The character
    following \ is to be known as the escaped character. An escape is a static
    substitution performed at parse time. If the escaped character is one of the
    following letters, abfnrtv, it has the standard meaning. If the escaped
    character is the asterisk, *, the parser erases all whitespace until the
    next nonwhitespace character. Otherwise, the parser ignores any meaning that
    it would otherwise attribute to the escaped character. In the string
    bracket, "", the only escapes that are processed are \* and \" and any other
    use of backslash is taken literally. If a quote bracket {} is taken
    literally, no escapes are processed but escaped {} do not count toward the
    requirement that all { must have a matching }.

LINE: A sequence of words terminated by newline, \n, or the semicolon, ;. The
    length of a line is the number of words it contains.

COMMENTS: A comment is a line whose first nonwhitespace character is the pound
    symbol #. To use # as the first character of a line literally it must be
    escaped. If it is not the first character of a line, there is no need to
    escape it. There must be bracketed {} in comments. This allows for multiline
    comments.

WORD: A bracket or character sequence delimited by whitespace or juxtapostion
    with a bracket. Any word may be prefixed by a sigil. Any special
    character to be literally included in a word must be escaped. Any UTF-8
    character that is not a special character or whitespace may be in the
    character sequence comprising the word without escaping. A literal word
    is a word that does not have a sigil and is not a bracket. If the word is
    not a bracket or sigil, the length of a word is defined as the number of
    characters the word comprises, after escaping is performed.

SIGIL: The splice, @, or the substitution, $, prefixing a word. The word
    prefixed by a sigil is said to be the word of that sigil. In a nonprefix
    position the sigil characters do not need to be escaped to be used as
    literals. The length of a sigil is defined as the length of the word it
    prefixes. A sigil must prefix a word.

BRACKET: A sequence of characters between an opening bracket, one of [{" and a
    closing bracket, one of "}]. A sequence between the [] brackets is a clause.
    A sequence between the {} brackets is a quote. A sequence between the ""
    brackets is a word that is delimited by " instead of whitespace, including
    newlines; no escapes are processed unless the escaped character is " itself
    or *. The length of a " bracketed word is the length of the characters
    between "" afer escaping. The only word of length 0 is represented by the
    literal "". Unbalanced, unescaped brackets are a syntax error.

CLAUSE: A clause is a word that is a line bracketed by []. The length of a
    clause is the length of the enclosed line.

QUOTE: A word that is a sequence of zero or more lines bracketed by {}.
    If a quote is used as a command, the usual escaping rules apply. If it is
    used as a literal word, no escaping is performed. Even if a quote is taken
    literally, all { must be matched by a properly nested }. Though escapes are
    not processed, escaped { or } do not count toward nesting. Though it need
    not be bracketed, the entirety of the input program is a quote. The length
    of a quote used literally is the length of the characters it contains, not
    including the outermost {}. A quote is a literal until it is used as code.


=EVALUATION=

    The words in a line are evaluated from left to right, and the line is
rewritten to the result of its evaluation.

    The lines of a quote are evaluated from top to bottom, and the quote is
rewritten to the result of evaluating its last line.

    Before a line is evaluated, it's words are rewritten:
        * A literal word or any quote is written as-is.
        * A substitution is rewritten as the result of dereferencing its word.
        * A splice dereferences its word. The result of dereferencing must be a
            list. Then, the list is expanded in place, increasing the number of
            arguments by one less the length of the list. If the list is empty,
            nothing is written, decreasing the number of arguments by one.
        * A clause is written as the result of executing its enclosed line.
            If either sigil prefixes the clause, it takes as input the result of
            rewriting the clause. If a clause is prefixed by a splice, the
            result of the clause must be a list.

    After rewriting its word, the first word in a line is the command and the
rest are the arguments of the command.
        * If the command is a quote or an alien, a command written in Go code,
            it is invoked with those arguments.
        * Otherwise, the command is a literal word that is dereferenced. The
            result of dereferencing must be either a quote or alien. If it is
            not a quote or alien, it is an error. It is then invoked as
            described above.
If the command is a quote, dereferencing the word arguments in its body returns
the arguments the callee invoked the quote with.
