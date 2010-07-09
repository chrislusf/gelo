package gelo

import "sync"

type _urw_mutex struct {
	rw sync.RWMutex
	u  sync.Mutex
}

func (m *_urw_mutex) Lock() {
	m.u.Lock()
	m.rw.Lock()
	m.u.Unlock()
}

func (m *_urw_mutex) Unlock() {
	m.rw.Unlock()
}

func (m *_urw_mutex) RLock() {
	m.u.Lock()
	m.rw.RLock()
	m.u.Unlock()
}

func (m *_urw_mutex) RUnlock() {
	m.rw.RUnlock()
}

func (m *_urw_mutex) Upgrade() {
	m.u.Lock()
	m.rw.RUnlock()
	m.rw.Lock()
	m.u.Unlock()
}

type namespace struct {
	up   *namespace
	dict *Dict
	mux  *_urw_mutex
}

func newNamespace(parent *namespace) *namespace {
	return &namespace{parent, NewDict(), &_urw_mutex{}}
}

func newNamespaceFrom(parent *namespace, dict *Dict) *namespace {
	return &namespace{parent, dict, &_urw_mutex{}}
}

func (ns *namespace) get(k Word) (Word, bool) {
	ns.mux.RLock()
	defer ns.mux.RUnlock()
	return ns.dict.Get(k)
}

func (ns *namespace) set(k, v Word) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	ns.dict.Set(k, v)
}

func (ns *namespace) del(k Word) {
	ns.mux.Lock()
	defer ns.mux.Unlock()
	ns.dict.Del(k)
}

//used to implement the (*VM).Read* family
func (ns *namespace) copyOut(s string) (w Word, ok bool) {
	for ; ns != nil; ns = ns.up {
		ns.mux.RLock()
		if w, ok = ns.dict.StrGet(s); ok {
			ns.mux.RUnlock()
			return w.DeepCopy(), true
		}
		ns.mux.RUnlock()
	}
	return nil, false
}

//another proxy like api to give namespace commands their own, well, namespace
type namespace_api struct {
	vm *VM
}

func (Ns *namespace_api) _is_blacklisted(s string) bool {
	h := Ns.vm.heritage
	if h == nil {
		return false
	}
	if h.blacklist == nil {
		return false
	}
	return h.blacklist[s]
}

func (Ns *namespace_api) Fork(n *namespace) {
	vm := Ns.vm
	if n == nil {
		vm.cns = newNamespace(vm.cns)
	} else {
		n.up = vm.cns
		vm.cns = n
	}
}

//returns false if we cannot Unfork (ie this is the topmost namespace)
func (Ns *namespace_api) Unfork() (*namespace, bool) {
	vm := Ns.vm
	if vm.cns.up == vm.top { //if no parent top = nil
		return nil, false
	}
	out := vm.cns
	vm.cns = out.up
	return out, true
}

func (Ns *namespace_api) Depth() (count int) {
	ns := Ns.vm.cns
	for ; ns != nil; ns = ns.up {
		count++
	}
	return
}

func (Ns *namespace_api) LocalDepth() (count int) {
	ns, top := Ns.vm.cns, Ns.vm.top
	//if top is nil this is exactly Depth()
	for ; ns != top; ns = ns.up {
		count++
	}
	return
}

func (Ns *namespace_api) DepthOf(name Word) (count int, there bool) {
	ns, top, str := Ns.vm.cns, Ns.vm.top, stringof(name.Ser())
	for ; ns != nil; ns = ns.up {
		if ns == top && Ns._is_blacklisted(str) {
			return
		}
		ns.mux.RLock()
		if ns.dict.StrHas(str) {
			ns.mux.RUnlock()
			there = true
			count++
			return
		}
		ns.mux.RUnlock()
		count++
	}
	return
}

func (Ns *namespace_api) Has(name Word) bool {
	ns, top, str := Ns.vm.cns, Ns.vm.top, stringof(name.Ser())
	for ; ns != nil; ns = ns.up {
		if ns == top && Ns._is_blacklisted(str) {
			return false
		}
		ns.mux.RLock()
		if ns.dict.StrHas(str) {
			ns.mux.RUnlock()
			return true
		}
		ns.mux.RUnlock()
	}
	return false
}

// lvls = 0 => only the current namespace
// lvls = n => first 1 + min(n, Depth()) namespaces
// lvls < 0 => capture Depth() namespaces
func (Ns *namespace_api) Locals(lvls int) *Dict {
	vm := Ns.vm
	ns, top, above := vm.cns, vm.top, false
	var blackl map[string]bool
	var count int
	var ok bool
	if vm.heritage != nil {
		blackl = vm.heritage.blacklist
	}
	m := make(map[string]Word)
	ns.mux.RLock()
	for k, v := range ns.dict.rep {
		m[k] = v
	}
	ns.mux.RUnlock()
	for ; count != lvls && ns != nil; ns = ns.up {
		count++
		above = above || ns == top //false until ns == top and true thereafter
		ns.mux.RLock()
		for k, v := range ns.dict.rep {
			if _, ok = m[k]; !ok {
				if above {
					if blackl != nil && blackl[k] {
						continue
					}
					v = v.DeepCopy()
				}
				m[k] = v
			}
		}
		ns.mux.RUnlock()
	}
	return &Dict{rep: m}
}

func (Ns *namespace_api) Lookup(name Word) (w Word, ok bool) {
	var above bool
	ns, top, str := Ns.vm.cns, Ns.vm.top, stringof(name.Ser())
	for ; ns != nil; ns = ns.up {
		if ns == top {
			above = true
			if Ns._is_blacklisted(str) {
				return nil, false
			}
		}
		ns.mux.RLock()
		if w, ok = ns.dict.StrGet(str); ok {
			ns.mux.RUnlock()
			if above {
				w = w.DeepCopy()
			}
			return w, true
		}
		ns.mux.RUnlock()
	}
	return nil, false
}

func (Ns *namespace_api) LookupOrElse(name Word) Word {
	w, ok := Ns.Lookup(name)
	if !ok {
		VariableUndefined(Ns.vm, name)
	}
	return w
}

func (Ns *namespace_api) Get(k Word) (Word, bool) {
	return Ns.vm.cns.get(k)
}

func (Ns *namespace_api) Set(k, v Word) {
	Ns.vm.cns.set(k, v)
}

func (Ns *namespace_api) _nthlvl(lvl int) (*namespace, bool) {
	ns, top := Ns.vm.cns, Ns.vm.top
	if lvl == 0 {
		return ns, true
	}
	for count := 0; ns.up != top; ns = ns.up {
		count++
		if count == lvl {
			return ns, true
		}
	}
	//reached top, but count != lvl
	if lvl < 0 {
		return ns, true
	}
	return nil, false
}

//returns false if lvl does not exist or we do not have write access
//if lvl is less than 0, write to the topmost namespace
func (Ns *namespace_api) NSet(lvl int, k, v Word) bool {
	ns, ok := Ns._nthlvl(lvl)
	if !ok {
		return false
	}
	ns.set(k, v)
	return true
}

//It is up to the caller to ensure that the incoming *Dict is not being
//written to by another goroutine during the run of Inject. Does no copying.
func (Ns *namespace_api) Inject(d *Dict) {
	ns := Ns.vm.cns
	t := ns.dict.rep
	ns.mux.Lock()
	defer ns.mux.Unlock()
	for k, v := range d.rep {
		t[k] = v
	}
}

func (Ns *namespace_api) NInject(lvl int, d *Dict) bool {
	ns, ok := Ns._nthlvl(lvl)
	if !ok {
		return false
	}
	t := ns.dict.rep
	ns.mux.Lock()
	defer ns.mux.Unlock()
	for k, v := range d.rep {
		t[k] = v
	}
	return true
}

func (Ns *namespace_api) Del(name Word) (Word, bool) {
	ns, top, str := Ns.vm.cns, Ns.vm.top, stringof(name.Ser())
	for ; ns != nil; ns = ns.up {
		if ns == top {
			//we do not blacklist unless it's already there
			for ; ns != nil; ns = ns.up {
				ns.mux.RLock()
				if v, ok := ns.dict.StrGet(str); ok {
					ns.mux.RUnlock()
					//blacklist
					h := Ns.vm.heritage
					if h.blacklist == nil {
						h.blacklist = make(map[string]bool)
					}
					h.blacklist[s] = true
					return v, true
				}
				ns.mux.RUnlock()
			}
			return nil, false
		}
		ns.mux.RLock()
		if v, ok := ns.dict.StrGet(str); ok {
			ns.mux.Upgrade()
			defer ns.mux.Unlock()
			ns.dict.StrDel(str)
			return v, true
		}
		ns.mux.RUnlock()
	}
	return nil, false
}

/*
 * Change value of s in place, if vm owns s, or assigns s with changed value
 * of s to vm's least deep ns if the vm doesn't own s. Returns false if
 * s is undefined in any ns. Deadlocks if the transformation function
 * attempts to access any of the locked namespaces, safest to not touch
 * any namespaces in the processes.
 */
func (Ns *namespace_api) MutateBy(name Word, f func(Word) (Word, bool)) (Word, bool) {
	ns, top, str := Ns.vm.cns, Ns.vm.top, stringof(name.Ser())
	var below *namespace
	above := false
	for ; ns != nil; ns = ns.up {
		if ns != top && !above {
			below = ns
		} else if ns == top {
			if Ns._is_blacklisted(str) {
				return nil, false
			}
			above = true
			//we are going to write to it soon and don't want anyone screwing
			//it up before we find the original s
			below.mux.Lock()
			defer below.mux.Unlock()
		}
		ns.mux.RLock()
		if w, ok := ns.dict.StrGet(str); ok {
			if above { //in a ns we don't own (target ns still locked)
				w = w.DeepCopy() //in case mutate alters w
				ns.mux.RUnlock()
				ns = below //already write locked
			} else { //if a ns we own
				ns.mux.Upgrade()
				defer ns.mux.Unlock()
			}
			if new, ok := f(w); ok {
				ns.dict.StrSet(str, new)
			}
			return w, true
		}
		ns.mux.RUnlock()
	}
	return nil, false
}

// Change value of s to w in original ns of s, or least deep ns if s is not
// owned by invoking VM.
func (Ns *namespace_api) Mutate(name, w Word) bool {
	ns, top, str := Ns.vm.cns, Ns.vm.top, stringof(name.Ser())
	var below *namespace
	above := false
	for ; ns != nil; ns = ns.up {
		if ns != top && !above {
			below = ns
		} else if ns == top {
			if Ns._is_blacklisted(str) {
				return false
			}
			above = true
			below.mux.Lock()
			defer below.mux.Unlock()
		}
		ns.mux.RLock()
		if old, ok := ns.dict.StrGet(str); ok {
			if above { //in a ns we don't own
				w = old.DeepCopy()
				ns.mux.RUnlock()
				ns = below
			} else {
				ns.mux.Upgrade()
				defer ns.mux.Unlock()
			}
			ns.dict.StrSet(str, w)
			return true
		}
		ns.mux.RUnlock()
	}
	return false
}

func (Ns *namespace_api) Swap(n1, n2 Word) (w1, w2 Word, ok bool) {
	//This operation becomes convoluted with this locking protocol
	//but it only has to be implemented once here and is relatively uncommon
	//operation
	ns, top := Ns.vm.cns, Ns.vm.top
	str1, str2 := stringof(n1.Ser()), stringof(n2.Ser())
	//lset and rset are only true in the iteration that they're found
	above, lset, rset := false, false, false
	var left, right, below *namespace
	for ; ns != nil; ns = ns.up {
		if ns != top && !above {
			below = ns
		} else if ns == top {
			if Ns._is_blacklisted(str1) || Ns._is_blacklisted(str2) {
				break
			}
			above = true
			//at least one symbol hasn't been found yet
			if left == nil || right == nil {
				below.mux.Lock()
				defer below.mux.Unlock()
			}
		}

		ns.mux.RLock()
		if left == nil {
			if w1, ok = ns.dict.StrGet(str1); ok {
				if above {
					lset = true
					w1 = w1.DeepCopy()
					left = below
				} else {
					left = ns
				}
			}
		}
		if right == nil {
			if w2, ok = ns.dict.StrGet(str2); ok {
				if above {
					rset = true
					w2 = w2.DeepCopy()
					right = below
				} else {
					right = ns
				}
			}
		}

		/* -If we are above, below is already locked
		 * -If we find either in top or above, we write to below and that lock
		 * is taken.
		 * -If one has been set this iteration and the other is unset, we
		 * upgrade the lock and continue looking
		 * -If both have been set this iteration, we upgrade the current lock
		 * -Otherwise we just release the read lock on the current ns
		 */
		if !above && lset {
			lset = false
			left.mux.Upgrade()
			defer left.mux.Unlock()
			if right != nil { //both found
				break
			}
		} else if !above && rset {
			rset = false
			right.mux.Upgrade()
			defer right.mux.Unlock()
			if left != nil {
				break
			}
		} else if !above && lset && rset {
			ns.mux.Upgrade()
			defer ns.mux.Unlock()
			break
		} else {
			ns.mux.RUnlock()
		}

		//above and both found
		if left != nil && right != nil {
			break
		}
	}

	//one or both not found
	if left == nil || right == nil {
		return nil, nil, false
	}

	//actually get to swap after all that
	left.dict.StrSet(str1, w2)
	right.dict.StrSet(str2, w1)
	return w2, w1, true
}
