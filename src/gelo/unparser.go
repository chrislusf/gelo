package gelo

func _unparse_line(code *sNode, buf *buffer) {
    for c := code; c != nil; c = c.next {
        switch c.tag {
            case synLiteral:
                buf.WriteWord(c.val.(Word))
            case synSplice, synIndirect:
                switch c.tag {
                    case synSplice:
                        buf.WriteString("@")
                    case synIndirect:
                        buf.WriteString("$")
                }
                _unparse_line(c.val.(*sNode), buf)
            case synClause:
                _unparse_line(c, buf)
            case synQuote:
                buf.Write(unparse(c.val.(Quote)))
        }
        buf.WriteString(" ")
    }
}

func unparse(q Quote) []byte {
    c, ok := q.unprotect().fcode()
    if !ok {
        //quote does not contain valid code, treat as string
        return q.Ser().Bytes()
    }
    buf := newBuf(0)
    line := c
    for ; line.next != nil; line = line.next {
        _unparse_line(line.cmd, buf)
        buf.WriteString("\n")
    }
    //handle 1 liners
    if line.next != nil {
        line = line.next
    }
    _unparse_line(line.cmd, buf)
    return buf.Bytes()
}

