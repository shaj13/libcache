package lfu

import (
	"container/heap"

	"github.com/shaj13/memc"
	"github.com/shaj13/memc/internal"
)

func init() {
	memc.LFU.Register(New)
}

// New returns new thread unsafe cache container.
func New(cap int) memc.Cache {
	f := &frequently{}
	f.Init()
	return internal.New(f, cap)
}

type element struct {
	value *internal.Entry
	index int
	count int
}

type frequently []*element

func (f frequently) Len() int {
	return len(f)
}

func (f frequently) Less(i, j int) bool {
	return f[i].count < f[j].count
}

func (f frequently) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
	f[i].index = i
	f[j].index = j
}

func (f *frequently) Push(v interface{}) {
	e := v.(*element)
	e.index = f.Len()
	*f = append(*f, e)
}

func (f *frequently) Pop() interface{} {
	e := (*f)[f.Len()-1]
	*f = (*f)[:f.Len()-1]
	return e
}

func (f *frequently) GetOldest() (e *internal.Entry) {
	return heap.Pop(f).(*element).value
}

func (f *frequently) Move(e *internal.Entry) {
	ele := e.Element.(*element)
	ele.count++
	heap.Fix(f, ele.index)
}

func (f *frequently) Remove(e *internal.Entry) {
	if e.Element.(*element).index < f.Len() {
		heap.Remove(f, e.Element.(*element).index)
	}
}

func (f *frequently) Add(e *internal.Entry) {
	ele := new(element)
	ele.value = e
	e.Element = ele
	heap.Push(f, ele)
}

func (f *frequently) Init() {
	*f = frequently{}
	heap.Init(f)
}
