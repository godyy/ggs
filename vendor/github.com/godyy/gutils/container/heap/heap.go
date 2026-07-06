package heap

import "container/heap"

// Element 实现Element即可作为Heap元素
type Element[E any] interface {
	// HeapLess Element比较
	HeapLess(E) bool

	// SetHeapIndex 设置索引
	SetHeapIndex(int)

	// HeapIndex 获取索引
	HeapIndex() int
}

// heapList 堆元素列表
// 实现heap.Interface的容器
type heapList[Elem Element[Elem]] []Elem

func (h heapList[Elem]) Len() int {
	return len(h)
}

func (h heapList[Elem]) Less(i, j int) bool {
	return h[i].HeapLess(h[j])
}

func (h heapList[Elem]) Swap(i, j int) {
	h[i].SetHeapIndex(j)
	h[j].SetHeapIndex(i)
	h[i], h[j] = h[j], h[i]
}

func (h *heapList[Elem]) Push(x any) {
	e := x.(Elem)
	e.SetHeapIndex(h.Len())
	*h = append(*h, e)
}

func (h *heapList[Elem]) Pop() any {
	n := h.Len() - 1
	e := (*h)[n]
	e.SetHeapIndex(-1)
	*h = (*h)[:n]
	return e
}

// Heap 堆容器
type Heap[Elem Element[Elem]] struct {
	minCap int            // 堆的最小容量
	list   heapList[Elem] // 堆元素列表
}

func NewHeap[Elem Element[Elem]](minCap ...int) *Heap[Elem] {
	h := &Heap[Elem]{}
	if len(minCap) > 0 && minCap[0] > 0 {
		h.minCap = minCap[0]
	}
	h.list = make(heapList[Elem], 0, h.minCap)
	return h
}

// Init 使用所给元素初始化堆
func (h *Heap[Elem]) Init(e ...Elem) {
	c := len(e)
	if c < h.minCap {
		c = h.minCap
	}
	h.list = make(heapList[Elem], 0, c)
	for _, v := range e {
		h.list.Push(v)
	}
	heap.Init(&h.list)
}

// Len 获取堆的当前长度
func (h *Heap[Elem]) Len() int {
	return h.list.Len()
}

// Push 元素入堆
func (h *Heap[Elem]) Push(e Elem) {
	if e.HeapIndex() >= 0 {
		panic("Element in other Heap")
	}
	heap.Push(&h.list, e)
}

// Pop 堆顶元素出堆
func (h *Heap[Elem]) Pop() Elem {
	if h.list.Len() <= 0 {
		panic("empty Heap")
	}
	x := heap.Pop(&h.list)
	h.adjustCap()
	return x.(Elem)
}

// Remove 移除堆中指定位置的元素
func (h *Heap[Elem]) Remove(i int) Elem {
	if i < 0 || i >= h.list.Len() {
		panic("index out of range")
	}
	x := heap.Remove(&h.list, i)
	h.adjustCap()
	return x.(Elem)
}

// Fix 修正指定位置元素的位置
func (h *Heap[Elem]) Fix(i int) {
	if i < 0 || i >= h.list.Len() {
		panic("index out of range")
	}
	heap.Fix(&h.list, i)
}

// Top 返回堆顶元素
func (h *Heap[Elem]) Top() Elem {
	if h.list.Len() <= 0 {
		panic("empty Heap")
	}
	return h.list[0]
}

// adjustCap 调整堆容量
func (h *Heap[Elem]) adjustCap() {
	c := cap(h.list)
	if c <= h.minCap {
		return
	}
	n := h.list.Len()
	if n < c/4 {
		c = c / 2
		if c < h.minCap {
			c = h.minCap
		}
		list := make(heapList[Elem], n, c)
		copy(list, h.list)
		h.list = list
	}
}
