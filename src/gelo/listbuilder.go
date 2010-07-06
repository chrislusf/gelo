package gelo

type _builder struct {
    head, tail  *List
}

//Create a new ListBuilder. The start parameter specifies any initial elements
func ListBuilder(start ...Word) *_builder {
    b := new(_builder)
    for _, v := range start {
        b._add(v)
    }
    return b
}

//Return the constructed List. It is not safe to use the ListBuilder after
//calling this.
func (b *_builder) List() *List {
    return b.head
}

func (b *_builder) _add(w Word) {
    cell := &List{w, nil}
    if b.head != nil {
        b.tail.Next = cell
        b.tail = b.tail.Next
    } else {
        b.head = cell
        b.tail = b.head
    }
}

func (b *_builder) Push(w Word) {
    b._add(w)
}

func (b *_builder) PushFront(w Word) {
    if b.head != nil {
        b.head = &List{w, b.head}
    } else {
        b.head = &List{w, nil}
    }
}

func (b *_builder) Extend(l *List) {
    for ; l != nil; l = l.Next {
        b._add(l.Value)
    }
}

func (b *_builder) ExtendFront(l *List) {
    if l == nil {
        return
    }
    middle := b.head
    b.head = l
    for ; l.Next != nil; l = l.Next {}
    l.Next = middle
}
