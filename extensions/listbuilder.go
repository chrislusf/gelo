package extensions

import "code.google.com/p/gelo"

//This is not intended to be used directly and is only exposed so that it can
//be stored in other types
type LBuilder struct {
	n          int
	head, tail *gelo.List
}

//Create a new ListBuilder. The start parameter specifies any initial elements
func ListBuilder(start ...gelo.Word) *LBuilder {
	b := new(LBuilder)
	for _, v := range start {
		b._add(v)
	}
	return b
}

//Call if you're going to create a closure in an alien using listbuilder but
//don't want to keep this around. Otherwise, it can be safely ignored.
func (b *LBuilder) Destroy() {
	b.head, b.tail = nil, nil
}

//Return the constructed List. It is not safe to use the ListBuilder after
//calling this.
func (b *LBuilder) List() *gelo.List {
	return b.head
}

func (b *LBuilder) _add(w gelo.Word) {
	cell := &gelo.List{w, nil}
	if b.head != nil {
		b.tail.Next = cell
		b.tail = b.tail.Next
	} else {
		b.head = cell
		b.tail = b.head
	}
	b.n++
}

func (b *LBuilder) Push(w gelo.Word) {
	b._add(w)
}

func (b *LBuilder) PushFront(w gelo.Word) {
	if b.head != nil {
		b.head = &gelo.List{w, b.head}
	} else {
		b.head = &gelo.List{w, nil}
	}
	b.n++
}

func (b *LBuilder) Extend(l *gelo.List) {
	for ; l != nil; l = l.Next {
		b._add(l.Value)
	}
}

func (b *LBuilder) ExtendFront(l *gelo.List) {
	if l == nil {
		return
	}
	middle := b.head
	b.head = l
	for ; l.Next != nil; l = l.Next {
		b.n++
	}
	l.Next = middle
}

func (b *LBuilder) Len() int {
	return b.n
}
