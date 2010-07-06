package gelo

import "sync"

type _urw_mutex struct {
    rw  sync.RWMutex
    u   sync.Mutex
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
    up     *namespace
    dict   *Dict
    mux    *_urw_mutex
}

func newNamespace(parent *namespace) *namespace {
    return &namespace{parent, NewDict(), &_urw_mutex{}}
}

func newNamespaceFrom(parent *namespace, dict *Dict) *namespace {
    return &namespace{parent, dict, &_urw_mutex{}}
}

func (ns *namespace) get(s Symbol) (Word, bool) {
    ns.mux.RLock()
    defer ns.mux.RUnlock()
    return ns.dict.Get(s)
}

func (ns *namespace) set(s Symbol, w Word) {
    ns.mux.Lock()
    defer ns.mux.Unlock()
    ns.dict.Set(s, w)
}

func (ns *namespace) del(s Symbol) {
    ns.mux.Lock()
    defer ns.mux.Unlock()
    ns.dict.Del(s)
}

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
type namespace_api struct { vm *VM }

func (Ns *namespace_api) Depth() (count byte) {
    ns := Ns.vm.cns
    for ; ns != nil; ns = ns.up {
        count++
    }
    return
}

func (Ns *namespace_api) LocalDepth() (count byte) {
    ns, top := Ns.vm.cns, Ns.vm.top
    //if top is nil this is exactly Depth()
    for ; ns != top; ns = ns.up {
        count++
    }
    return
}

func (Ns *namespace_api) DepthOf(s Symbol) (count byte, there bool) {
    str := s.String()
    for ns := Ns.vm.cns; ns != nil; ns = ns.up {
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

func (Ns *namespace_api) Has(s Symbol) bool {
    ns, str := Ns.vm.cns, s.String()
    for ; ns != nil; ns = ns.up {
        ns.mux.RLock()
        if ns.dict.StrHas(str) {
            ns.mux.RUnlock()
            return true
        }
        ns.mux.RUnlock()
    }
    return false
}

func (Ns *namespace_api) Lookup(s Symbol) (w Word, ok bool) {
    ns, str := Ns.vm.cns, s.String()
    for ; ns != nil; ns = ns.up {
        ns.mux.RLock()
        if w, ok = ns.dict.StrGet(str); ok {
            ns.mux.RUnlock()
            return w, true
        }
        ns.mux.RUnlock()
    }
    return nil, false
}

func (Ns *namespace_api) LookupOrElse(s Symbol) Word {
    w, ok := Ns.Lookup(s)
    if !ok {
        VariableUndefined(Ns.vm, s)
    }
    return w
}

func (Ns *namespace_api) Set(s Symbol, w Word) {
    Ns.vm.cns.set(s, w)
}

func (Ns *namespace_api) Del(s Symbol) (Word, bool) {
    ns, top, str := Ns.vm.cns, Ns.vm.top, s.String()
    above := false
    for ; ns != nil; ns = ns.up {
        if ns == top {
            above = true
        }
        ns.mux.RLock()
        if v, ok := ns.dict.StrGet(str); ok {
            if above {
                ns.mux.RUnlock()
                RuntimeError(Ns.vm, "Access violation. Attempted to delete",
                    s, "which belongs to a parent vm")
                    
            }
            ns.mux.Upgrade()
            defer ns.mux.Unlock()
            ns.dict.StrDel(str)
            return v, true
        }
        ns.mux.RUnlock()
    }
    return Null, false
}

/*
 * Change value of s in place, if vm owns s, or assigns s with changed value
 * of s to vm's least deep ns if the vm doesn't own s. Returns false if
 * s is undefined in any ns. Deadlocks if the transformation function
 * attempts to access any of the locked namespaces, safest to not touch
 * any namespaces in the processes.
 */ 
func (Ns *namespace_api) MutateBy(s Symbol, f func(Word)(Word,bool)) (Word, bool) {
    ns, top, str := Ns.vm.cns, Ns.vm.top, s.String()
    var below *namespace
    above := false
    for ; ns != nil; ns = ns.up {
        if ns != top && !above {
            below = ns
        } else if ns == top {
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
    return Null, false
}

// Change value of s to w in original ns of s, or least deep ns if s is not
// owned by invoking VM.
func (Ns *namespace_api) Mutate(s Symbol, w Word) bool {
    ns, top, str := Ns.vm.cns, Ns.vm.top, s.String()
    var below *namespace
    above := false
    for ; ns != nil; ns = ns.up {
        if ns != top && !above {
            below = ns
        } else if ns == top {
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

func (Ns *namespace_api) Swap(s1, s2 Symbol) (w1, w2 Word, ok bool) {
    //This operation becomes convoluted with this locking protocol
    //but it only has to be implemented once here and is relatively uncommon
    //operation
    ns, top, str1, str2 := Ns.vm.cns, Ns.vm.top, s1.String(), s2.String()
    //lset and rset are only true in the iteration that they're found
    above, lset, rset := false, false, false
    var left, right, below *namespace

    for ; ns != nil; ns = ns.up {
        if ns != top && !above {
            below = ns
        } else if ns == top {
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
        return Null, Null, false
    }

    //actually get to swap after all that
    left.dict.StrSet(str1, w2)
    right.dict.StrSet(str2, w1)
    return w2, w1, true
}
