package process

import (
	"fmt"
	"sort"

	mes "github.com/qwertyqq2/entity/message"
)

type Wantlist struct {
	set   map[string]Entry
	cache []Entry
}

type Entry struct {
	key       string
	op        mes.Op
	data      mes.Inside
	undefined bool
}

func (e Entry) Empty() bool {
	if e.data.Data == "" {
	} else {
		return false
	}
	switch e.op {
	case mes.Add, mes.Get, mes.Has, mes.Delete, mes.Fail, mes.Success, mes.Ping:
		return false
	default:
		return true
	}
}

func (e Entry) String() string {
	return fmt.Sprintf("op: %d, key: %s, data: %s\n", e.op, e.data.Key, e.data.Data)
}

type entrySlice []Entry

func (es entrySlice) Len() int           { return len(es) }
func (es entrySlice) Swap(i, j int)      { es[i], es[j] = es[j], es[i] }
func (es entrySlice) Less(i, j int) bool { return len(es[i].data.Data) > len(es[j].data.Data) }

func NewWantlist() *Wantlist {
	return &Wantlist{
		set: make(map[string]Entry),
	}
}

func (w *Wantlist) Len() int {
	return len(w.set)
}

func (w *Wantlist) Add(key string, op mes.Op, data mes.Inside) bool {
	e, ok := w.set[key]

	if ok && e.data.Key != "" {
		return false
	}

	w.put(key, Entry{key: key, op: op, data: data})
	return true
}

func (w *Wantlist) Remove(key string) bool {
	if _, ok := w.set[key]; !ok {
		return false
	}
	w.delete(key)
	return true
}

func (w *Wantlist) Contains(key string) (Entry, bool) {
	e, ok := w.set[key]
	return e, ok
}

func (w *Wantlist) Entries() []Entry {
	if w.cache != nil {
		return w.cache
	}
	es := make([]Entry, 0, len(w.set))
	for _, e := range w.set {
		es = append(es, e)
	}
	sort.Sort(entrySlice(es))
	w.cache = es
	return es[0:len(es):len(es)]
}

func (w *Wantlist) Absorf(other *Wantlist) {
	w.cache = nil
	for _, e := range other.Entries() {
		w.Add(e.key, e.op, e.data)
	}
}

func (w *Wantlist) String() string {
	res := ""
	for k, d := range w.set {
		res += fmt.Sprintf("op: %d, key: %s, data: %s \n", d.op, k, d.data.Data)
	}
	return res
}

func (w *Wantlist) put(k string, e Entry) {
	w.cache = nil
	w.set[k] = e
}

func (w *Wantlist) delete(k string) {
	delete(w.set, k)
	w.cache = nil
}
