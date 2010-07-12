package extensions

import "gelo"

type _builder struct {
	head, tail *gelo.List
}

//Create a new ListBuilder. The start parameter specifies any initial elements
func ListBuilder(start ...gelo.Word) *_builder {
	b := new(_builder)
	for _, v := range start {
		b._add(v)
	}
	return b
}

//Call if you're going to create a closure in an alien using listbuilder but
//don't want to keep this around. Otherwise, it can be safely ignored.
func (b *_builder) Destroy() {
	b.head, b.tail = nil, nil
}

//Return the constructed List. It is not safe to use the ListBuilder after
//calling this.
func (b *_builder) List() *gelo.List {
	return b.head
}

func (b *_builder) _add(w gelo.Word) {
	cell := &gelo.List{w, nil}
	if b.head != nil {
		b.tail.Next = cell
		b.tail = b.tail.Next
	} else {
		b.head = cell
		b.tail = b.head
	}
}

func (b *_builder) Push(w gelo.Word) {
	b._add(w)
}

func (b *_builder) PushFront(w gelo.Word) {
	if b.head != nil {
		b.head = &gelo.List{w, b.head}
	} else {
		b.head = &gelo.List{w, nil}
	}
}

func (b *_builder) Extend(l *gelo.List) {
	for ; l != nil; l = l.Next {
		b._add(l.Value)
	}
}

func (b *_builder) ExtendFront(l *gelo.List) {
	if l == nil {
		return
	}
	middle := b.head
	b.head = l
	for ; l.Next != nil; l = l.Next {
	}
	l.Next = middle
}
