package commands

import (
	"gelo"
	"gelo/extensions"
	"math"
	"bytes"
	"sort"
)

func ListCon(_ *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 1 {
		l, ok := gelo.UnserializeListFrom(args.Value)
		if ok {
			return l
		}
	}
	return args
}

var _make_list_parser = extensions.MakeOrElseArgParser(
	"length 'long 'with zero-value")

func Make_list(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	Args := _make_list_parser(vm, args)
	length, ok := vm.API.NumberOrElse(Args["length"]).Int()
	if !ok || length < 1 {
		gelo.TypeMismatch(vm, "nonzero positive integer", "number")
	}
	zero, list := Args["zero-value"], extensions.ListBuilder()
	for i := int64(1); i < length; i++ {
		list.Push(zero.Copy())
	}
	return list.List()
}

// arg-count returns the number of arguments that it is called with.
// It is of limited functionality but is occasionally of use
func Arg_count(_ *gelo.VM, _ *gelo.List, ac uint) gelo.Word {
	n, _ := gelo.NewNumberFromGo(ac)
	return n
}

func LLength(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "llength", "list+", args)
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		n, _ := gelo.NewNumberFromGo(vm.API.ListOrElse(w).Len())
		return n
	})
}

var _partition_parser = extensions.MakeOrElseArgParser("list 'by command")

func Partition(vm *gelo.VM, args *gelo.List, _ uint) gelo.Word {
	Args := _partition_parser(vm, args)
	list := vm.API.ListOrElse(Args["list"])
	cmd := vm.API.InvokableOrElse(Args["command"])
	buckets := make(map[string]*extensions.LBuilder)
	for ; list != nil; list = list.Next {
		key := vm.API.InvokeCmdOrElse(cmd, gelo.AsList(list.Value)).Ser().String()
		if bt, there := buckets[key]; there {
			bt.Push(list.Value)
		} else { //first item in this class
			buckets[key] = extensions.ListBuilder(list.Value)
		}
	}
	builder := extensions.ListBuilder()
	for _, v := range buckets {
		builder.Push(v.List())
	}
	return builder.List()
}

func Head(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "head", "list+", "")
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		l := vm.API.ListOrElse(w)
		if l == nil {
			return gelo.Null
		}
		return l.Value
	})
}

func Tail(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "tail", "list+", "")
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		l := vm.API.ListOrElse(w)
		if l == nil {
			return gelo.EmptyList
		}
		return l.Next
	})
}

func LIndex(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac < 2 {
		gelo.ArgumentError(vm, "lindex", "list indicies+", args)
	}
	l := vm.API.ListOrElse(args.Value)
	list := l.Slice()
	if ac == 2 { //only one index
		return list[ToIdx(vm, args.Next.Value, len(list))]
	}
	idxs := make([]int, ac-1)
	count := 0
	for i := args.Next; i != nil; i = i.Next {
		idxs[count] = ToIdx(vm, i.Value, len(list))
		count++
	}
	head := &gelo.List{list[idxs[0]], nil}
	tail := head
	for _, v := range idxs[1:] {
		tail.Next = &gelo.List{list[v], nil}
		tail = tail.Next
	}
	return head
}

func Zip(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "zip", "list+", args)
	}
	//set up ring of args, typecheck
	var _rhead, ring *gelo.List
	length := 0
	for ; args != nil; args = args.Next {
		l := vm.API.ListOrElse(args.Value)
		if l == nil {
			return gelo.EmptyList
		}
		length++
		if _rhead != nil {
			ring.Next = &gelo.List{l, nil}
			ring = ring.Next
		} else {
			_rhead = &gelo.List{l, nil}
			ring = _rhead
		}
	}
	ring.Next = _rhead
	//build zipped list of lists
	var head, tail, curh, curt *gelo.List
	for {
		for i := 0; i < length; i++ {
			ring = ring.Next
			node := ring.Value.(*gelo.List)
			if node == nil {
				ring = nil //break cycle
				return head
			}
			v := node.Value
			if curh != nil {
				curt.Next = &gelo.List{v, nil}
				curt = curt.Next
			} else {
				curh = &gelo.List{v, nil}
				curt = curh
			}
			ring.Value = node.Next
		}
		if head != nil {
			tail.Next = &gelo.List{curh, nil}
			tail = tail.Next
		} else {
			head = &gelo.List{curh, nil}
			tail = head
		}
		curh = nil //reset curh
	}
	panic("zip in impossible state") //Issue 65
}

var _rparser = extensions.MakeOrElseArgParser("[a 'to]? b ['by i]?")

func _rassnum(vm *gelo.VM, d map[string]gelo.Word, key string) (float64, bool) {
	w, ok := d[key]
	if !ok {
		return 0, false
	}
	n := vm.API.NumberOrElse(w)
	return n.Real(), true
}
func Range(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	d := _rparser(vm, args)
	a, start := _rassnum(vm, d, "a")
	b, _ := _rassnum(vm, d, "b")
	i, step := _rassnum(vm, d, "i")
	//XXX +n by -m should either error or have a sane default behavior
	if !start {
		a = 0
	}
	if !step {
		if a < b {
			i = 1
		} else {
			i = -1
		}
	}
	if i == 0 {
		gelo.RuntimeError(vm, "range step size cannot be 0")
	}
	if math.Fabs(b-a) < math.Fabs(i) {
		n := gelo.NewNumber(0)
		return gelo.NewList(n)
	}
	n, _ := gelo.NewNumberFromGo(a)
	list := extensions.ListBuilder(n)
	var cmp func(a, b float64) bool
	if a < b {
		cmp = func(a, b float64) bool { return a < b }
	} else {
		cmp = func(a, b float64) bool { return b < a }
	}
	for m := a + i; cmp(m, b); m += i {
		n, _ := gelo.NewNumberFromGo(m)
		list.Push(n)
	}
	return list.List()
}

func LReverse(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "lreverse", "list", args)
	}
	in, out := vm.API.ListOrElse(args.Value), extensions.ListBuilder()
	for ; in != nil; in = in.Next {
		out.PushFront(in.Value)
	}
	return out.List()
}

func Unique(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "uniq", "list", args)
	}
	l, list := vm.API.ListOrElse(args.Value), extensions.ListBuilder()
	for seen := make(map[string]bool); l != nil; l = l.Next {
		item := l.Value.Ser().String()
		if !seen[item] {
			seen[item] = true
			list.Push(l.Value)
		}
	}
	return list.List()
}

//index-of value list
//returns a list of the indicies of value within list
func Index_of(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "index-of", "value list", args)
	}
	val := args.Value
	l, list := vm.API.ListOrElse(args.Next.Value), extensions.ListBuilder()
	for count := 0; l != nil; l, count = l.Next, count + 1 {
		if val.Equals(l.Value) {
			list.Push(gelo.NewNumber(float64(count)))
		}
	}
	return list.List() //may be the empty list
}

func Enumerate(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "enumerate", "list", args)
	}
	list, builder := vm.API.ListOrElse(args.Value), extensions.ListBuilder()
	for count := 0; list != nil; list, count = list.Next, count + 1 {
		n, _ := gelo.NewNumberFromGo(count)
		builder.Push(gelo.NewList(gelo.NewList(n, list.Value)))
	}
	return builder.List()
}

var _every_parser = extensions.MakeOrElseArgParser(
	"['item name 'in]? list 'do command")

func Every(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	Args := _every_parser(vm, args)
	list := vm.API.ListOrElse(Args["list"])
	cmd := Args["command"]
	if !possiblyInvokable(cmd) {
		gelo.TypeMismatch(vm, "invokable", cmd.Type())
	}
	_, named := Args["item"]
	var name gelo.Word
	if named {
		name = Args["name"]
		if d, there := vm.Ns.DepthOf(name); there && d == 0 {
			old := vm.Ns.LookupOrElse(name)
			defer vm.Ns.Set(0, name, old)
		} else if list != nil {
			defer vm.Ns.Del(name)
		}
	}
	return list.Map(func(w gelo.Word) gelo.Word {
		if named {
			vm.Ns.Set(0, name, w)
		}
		return vm.API.InvokeCmdOrElse(cmd, gelo.NewList(w))
	})
}

var _some_parser = extensions.MakeOrElseArgParser(
	"['item name 'in]? list 'by command")

func Some(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	Args := _some_parser(vm, args)
	list := vm.API.ListOrElse(Args["list"])
	_, named := Args["item"]
	var name gelo.Word
	if named {
		name = vm.API.SymbolOrElse(Args["name"])
		if d, there := vm.Ns.DepthOf(name); there && d == 0 {
			old := vm.Ns.LookupOrElse(name)
			defer vm.Ns.Set(0, name, old)
		} else if list != nil {
			defer vm.Ns.Del(name)
		}
	}
	cmd := Args["command"]
	if !possiblyInvokable(cmd) {
		gelo.TypeMismatch(vm, "invokable", cmd.Type())
	}
	var head, tail *gelo.List
	for ; list != nil; list = list.Next {
		v := list.Value
		if named {
			vm.Ns.Set(0, name, v)
		}
		val := vm.API.InvokeCmdOrElse(cmd, gelo.NewList(v))
		if b, ok := val.(gelo.Bool); ok && b.True() {
			if head != nil {
				tail.Next = &gelo.List{v, nil}
				tail = tail.Next
			} else {
				head = &gelo.List{v, nil}
				tail = head
			}
		}
	}
	return head
}

var _reduce_parser = extensions.MakeOrElseArgParser(
	"['initial value]? ['items x y 'in]? list 'with command")

func Reduce(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	Args := _reduce_parser(vm, args)
	list := vm.API.ListOrElse(Args["list"])
	cmd := Args["command"]
	if !possiblyInvokable(cmd) {
		gelo.TypeMismatch(vm, "invokable", cmd.Type())
	}
	_, named := Args["items"]
	var left, right gelo.Word
	if named {
		left, right = Args["x"], Args["y"]
		if d, there := vm.Ns.DepthOf(left); there && d == 0 {
			old := vm.Ns.LookupOrElse(left)
			defer vm.Ns.Set(0, left, old)
		} else if list != nil {
			defer vm.Ns.Del(left)
		}
		if d, there := vm.Ns.DepthOf(right); there && d == 0 {
			old := vm.Ns.LookupOrElse(right)
			defer vm.Ns.Set(0, right, old)
		} else if list != nil {
			defer vm.Ns.Del(right)
		}
	}
	var acc gelo.Word
	_, there := Args["initial"]
	if there {
		acc = Args["value"]
	} else {
		acc = list.Value
		list = list.Next
	}
	for ; list != nil; list = list.Next {
		v := list.Value
		if named {
			vm.Ns.Set(0, left, acc)
			vm.Ns.Set(0, right, v)
		}
		acc = vm.API.InvokeCmdOrElse(cmd, gelo.NewList(acc, v))
	}
	return acc
}

func Intersect(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		gelo.ArgumentError(vm, "intersect", "list list", args)
	}
	left := vm.API.ListOrElse(args.Value)
	right := vm.API.ListOrElse(args.Next.Value)
	if left == nil || right == nil {
		return gelo.EmptyList
	}
	var head, tail *gelo.List
	for ; left != nil; left = left.Next {
		v := left.Value
		for r := right; r != nil; r = r.Next {
			if v.Equals(r.Value) {
				if head != nil {
					tail.Next = &gelo.List{v, nil}
					tail = tail.Next
				} else {
					head = &gelo.List{v, nil}
					tail = head
				}
			}
		}
	}
	return head
}

var _comp_parser = extensions.MakeOrElseArgParser("list1 'wrt list2")

func Complement_of(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	Args := _comp_parser(vm, args)
	A := vm.API.ListOrElse(Args["list1"])
	B := vm.API.ListOrElse(Args["list2"])
	var head, tail *gelo.List
	seen := make(map[string]bool)
	for ; B != nil; B = B.Next {
		v := B.Value
		s := v.Ser().String()
		if seen[s] {
			continue
		}
		seen[s] = true
		for a := A; a != nil; a = a.Next {
			if v.Equals(a.Value) {
				break
			}
			if a.Next == nil { //made it through without hitting anything
				if head != nil {
					tail.Next = &gelo.List{v, nil}
					tail = tail.Next
				} else {
					head = &gelo.List{v, nil}
					tail = head
				}
			}
		}
	}
	return head
}

func Sym_diff(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "sym-diff", "list1 list2", args)
	}
	A := vm.API.ListOrElse(args.Value)
	B := vm.API.ListOrElse(args.Next.Value)
	if A == nil {
		return B
	}
	if B == nil {
		return A
	}
	Am := make(map[string]gelo.Word)
	Bm := make(map[string]gelo.Word)
	for ; A != nil; A = A.Next {
		v := A.Value
		Am[v.Ser().String()] = v
	}
	for ; B != nil; B = B.Next {
		v := B.Value
		Bm[v.Ser().String()] = v
	}
	var head, tail *gelo.List
	for k, v := range Am {
		if _, ok := Bm[k]; ok { //in both
			Bm[k] = v, false //delete
		} else {
			if head != nil {
				tail.Next = &gelo.List{v, nil}
				tail = tail.Next
			} else {
				head = &gelo.List{v, nil}
				tail = head
			}
		}
	}
	//we've removed everything in intersect(A, B)
	for _, v := range Bm {
		if head != nil {
			tail.Next = &gelo.List{v, nil}
			tail = tail.Next
		} else { //in case A was a subset of B
			head = &gelo.List{v, nil}
			tail = head
		}
	}
	return head
}

func Subseqp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "subseq?", "list1 list2", args)
	}
	A := vm.API.ListOrElse(args.Value).Slice()
	B := vm.API.ListOrElse(args.Next.Value).Slice()
	if len(A) < len(B) {
		return gelo.False
	}
	for i, j := 0, 0; j < len(A); i++ {
		if A[i].Equals(B[j]) {
			j++
			if j == len(B) {
				return gelo.True
			}
		} else {
			j = 0
			if len(A)-i < len(B) {
				return gelo.False
			}
		}
	}
	panic("Subseqp in impossible state") //Issue 65
}

func Subsetp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 2 {
		gelo.ArgumentError(vm, "subset?", "list1 list2", args)
	}
	A := vm.API.ListOrElse(args.Value)
	B := vm.API.ListOrElse(args.Next.Value)
	m := make(map[string]bool)
	for ; A != nil; A = A.Next {
		m[A.Value.Ser().String()] = true
	}
	for ; B != nil; B = B.Next {
		if !m[B.Value.Ser().String()] {
			return gelo.False
		}
	}
	return gelo.True
}

type _warray struct {
	cache  map[int][]byte
	perms  []int
	array  []gelo.Word
	length int
}

func (p *_warray) Len() int {
	return p.length
}
func (p *_warray) Swap(i, j int) {
	p.perms[i], p.perms[j] = p.perms[j], p.perms[i]
}
func (p *_warray) get(i int) []byte {
	if s, ok := p.cache[i]; ok {
		return s
	}
	s := p.array[i].Ser().Bytes()
	p.cache[i] = s
	return s
}
func (p *_warray) Less(i, j int) bool {
	return bytes.Compare(p.get(p.perms[i]), p.get(p.perms[j])) == -1
}
func LSort(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac != 1 {
		gelo.ArgumentError(vm, "lsort", "list", args)
	}
	o := vm.API.ListOrElse(args.Value).Slice()
	length := len(o)
	perms := make([]int, length)
	for i := 0; i < length; i++ {
		perms[i] = i
	}
	s := &_warray{make(map[int][]byte), perms, o, length}
	sort.Sort(s)
	fst, _ := gelo.NewNumberFromGo(s.perms[0])
	head := &gelo.List{fst, nil}
	tail := head
	for _, v := range s.perms[1:] {
		p, _ := gelo.NewNumberFromGo(v)
		tail.Next = &gelo.List{p, nil}
		tail = tail.Next
	}
	return head
}

func Empty_listp(vm *gelo.VM, args *gelo.List, ac uint) gelo.Word {
	if ac == 0 {
		return gelo.True
	}
	return args.MapOrApply(func(w gelo.Word) gelo.Word {
		return gelo.ToBool(vm.API.ListOrElse(w) == gelo.EmptyList)
	})
}

var ListCommands = map[string]interface{}{
	"List":          ListCon,
	"make-list":     Make_list,
	"partition":     Partition,
	"head":          Head,
	"tail":          Tail,
	"unique":        Unique,
	"index-of":      Index_of,
	"lindex":        LIndex,
	"arg-count":     Arg_count,
	"llength":       LLength,
	"lreverse":      LReverse,
	"enumerate":     Enumerate,
	"zip":           Zip,
	"range":         Range,
	"every":         Every,
	"some":          Some,
	"reduce":        Reduce,
	"intersect":     Intersect,
	"complement-of": Complement_of,
	"subseq?":       Subseqp,
	"subset?":       Subsetp,
	"sym-diff":      Sym_diff,
	"lsort":         LSort,
	"empty-list?":   Empty_listp,
}
