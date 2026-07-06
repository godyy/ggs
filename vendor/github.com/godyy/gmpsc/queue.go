// Package gmpsc 提供一个无锁的多生产者、单消费者（MPSC）队列实现。
//
// 设计要点：
// - 链式分段：队列由一条 segment 链组成，每段容量固定，生产者在尾段写满后协助链接下一段；
// - 位图就绪标记：生产者写入 data[idx] 后置 ready 位表明该槽可被消费；消费者先读位再读 data，避免读到未完成写入的槽；
// - 顺序保证：依赖原子操作的顺序一致性（SC）保证“写 data -> 置位 -> 读位 -> 读 data”的可见性次序；
// - 容量上限：cap 限制“在队元素数量”（已占用但尚未被消费/回收），用于限流与丢弃策略；
// - 复用内存：按元素类型与分段大小共享 sync.Pool，降低分段分配/回收成本；
// - 回收安全：分段通过 refs/dead 进行引用计数，避免消费者推进 head 后，仍在访问旧段的生产者误用已回收对象。
package gmpsc

import (
	"math/bits"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Queue 是一个 MPSC（多生产者、单消费者）队列：
// - Enqueue 可被多个 goroutine 并发调用
// - Dequeue 只能由单一 goroutine 调用
//
// 本实现采用链式分段（segment）+ 位图 ready 标记以降低内存占用，并通过 cap 对“在队元素数量”做硬上限限制。
type Queue[T any] struct {
	// cap 为在队元素数量上限；cap<=0 表示不限制。
	cap int64
	// size 为“已占用但尚未被消费”的槽数量，用于 cap 限流。
	// 注意：size 是并发近似值；在高并发下允许短暂偏差，但不会突破 cap 的硬上限。
	size atomic.Int64
	// ready 为“已写入且可被消费”的元素数量。
	// 与 size 的关系：ready <= size；两者差值对应“已占用但 ready 位尚未置位”的槽（生产者写入进行中）。
	ready atomic.Int64
	// closed 标记队列是否已关闭。
	closed atomic.Bool
	// activeOps 追踪当前正在执行的 Enqueue/Dequeue 操作数，用于 Close 时安全回收。
	activeOps atomic.Int32
	// head 仅由消费者 goroutine 读写：指向当前消费段。
	head *segment[T]
	// tail 由生产者并发推进：入队写满当前段后协助链接下一段并尝试推进 tail。
	tail atomic.Pointer[segment[T]]
	// segPool 复用分段对象；同 (T, segSize) 的队列可共享该池。
	segPool *sync.Pool
}

// segment 是队列中的一个固定大小分段：
// - 生产者仅写 data[idx] 并置位 ready（并发多写）；消费者仅按 r 顺序读取（单读）；
// - data：保存实际元素。生产者根据 w 的返回 idx 写入 data[idx]；
// - ready：位图；某位为 1 表示对应 data[idx] 已写入并对消费者可见；
// - w：写入序号（多生产者并发，原子自增分配 idx）；越界表示该段已写满；
// - r：读取序号（单消费者递增）；与 ready 位配合保证 FIFO；
// - next：指向下一段；生产者通过 CAS(nil, 新段) 协助链接；消费者持有 head 逐段前进。
// - refs/dead：分段回收控制。队列持有 1 个引用；生产者在使用分段前 acquire 增引用，使用后 releaseSeg 归还。
type segment[T any] struct {
	next  atomic.Pointer[segment[T]]
	data  []T
	ready []uint64 // bitset: 1 = ready, 0 = not ready
	w     atomic.Uint64
	r     uint64
	refs  atomic.Int32
	dead  atomic.Bool
}

// segKey 用于为不同的 Queue[T] 共享分段池：
// - typ：元素类型 T（不同 T 不能共享同一分段，因为 segment 内持有 []T）
// - size：分段大小（同一 T 但不同分段大小也不能混用）
type segKey struct {
	typ  reflect.Type
	size int
}

// segPools 存储按 (T, segSize) 分组的共享分段池。
// 共享池仅影响“分段对象”的复用，不改变队列的并发语义。
var segPools sync.Map // key: segKey -> *sync.Pool of *segment[T]

// getSharedSegPool 返回按 (T, size) 共享的分段池。
// reflect 仅发生在创建队列（调用 NewQueue）时的初始化路径，正常 Enqueue/Dequeue 路径不涉及反射或 sync.Map。
func getSharedSegPool[T any](size int) *sync.Pool {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	key := segKey{typ: typ, size: size}
	if p, ok := segPools.Load(key); ok {
		return p.(*sync.Pool)
	}
	pool := &sync.Pool{
		New: func() any { return newSegment[T](size) },
	}
	actual, _ := segPools.LoadOrStore(key, pool)
	return actual.(*sync.Pool)
}

// ceilPow2 将 n 归一化为不小于 n 的 2 的整数次幂（向上取整）。
func ceilPow2(n int) int {
	if n <= 1 {
		return 1
	}
	// 若已是 2 的幂，直接返回
	if n&(n-1) == 0 {
		return n
	}
	return 1 << bits.Len(uint(n))
}

// NewQueue 创建一个队列。
//
// 参数：
//   - segSize：每个分段的 data 长度；<=0 时使用默认值 512；创建时归一化为不小于 segSize 的 2 的幂；
//   - cap：在队元素数量上限；<0 视为 0；cap<=0 表示不限制。
func NewQueue[T any](segSize int, cap int) *Queue[T] {
	if segSize <= 0 {
		segSize = 512
	}
	segSize = ceilPow2(segSize)
	if cap < 0 {
		cap = 0
	}
	q := &Queue[T]{
		cap:     int64(cap),
		segPool: getSharedSegPool[T](segSize),
	}
	s := q.getSeg()
	q.initSeg(s)
	q.head = s
	q.tail.Store(s)
	return q
}

// relaxSpin 根据自旋次数进行自适应退让：
// - 前期纯自旋：避免打断低延迟路径；
// - 中期让出调度（Gosched）：缓和竞争、减少调度风暴；
// - 长时间高竞争：短暂 Sleep(1µs) 降低抖动。
// 用于 Enqueue 的高竞争阶段，平衡吞吐与尾延迟。
func relaxSpin(spins int) {
	if spins&0x3F != 0 {
		return
	}
	if spins < 1<<12 {
		runtime.Gosched()
		return
	}
	time.Sleep(time.Microsecond)
}

// Enqueue 尝试入队一个元素（多生产者安全）。
// 返回值：
// - ok：是否成功入队（达到 cap 上限则返回 false）；
//
// 并发语义：
// 1) 生产者先通过 tryAcquire 占用“在队配额”；
// 2) 在当前尾段通过原子自增分配写入槽 idx，先写 data[idx]，后置 ready 位；
// 3) 段写满时，生产者采用“协助式链接”：任意生产者都可 CAS 将 next 从 nil 链接到新段，并尝试推进 tail；
// 4) 在极端竞争下，采取轻量 Gosched 让出，降低忙等抖动。
func (q *Queue[T]) Enqueue(v T) bool {
	// 增加活动操作数，确保 Close 时安全回收
	q.activeOps.Add(1)
	defer q.activeOps.Add(-1)

	// 检查队列是否已关闭
	if q.closed.Load() {
		return false
	}

	if !q.tryAcquire() {
		return false
	}

	spins := 0

	for {
		if spins != 0 {
			relaxSpin(spins)
		}

		t := q.tail.Load()
		if t == nil {
			// 队列已关闭时 tail 可能被置为 nil，此时退出
			return false
		}

		// 尝试引用分段
		if !t.acquire() {
			spins++
			continue
		}

		// 尝试更新分段元素
		idx := t.w.Add(1) - 1
		if idx < uint64(len(t.data)) {
			t.data[idx] = v
			t.setReadyAt(idx)
			q.releaseSeg(t)
			q.ready.Add(1)
			return true
		}

		// 尝试新增分段
		n := t.next.Load()
		if n == nil {
			ns := q.getSeg()
			q.initSeg(ns)
			if t.next.CompareAndSwap(nil, ns) {
				n = ns
			} else {
				q.putSeg(ns)
				n = t.next.Load()
			}
		}

		// 尝试推进分段
		if n != nil {
			_ = q.tail.CompareAndSwap(t, n)
		}

		q.releaseSeg(t)
		spins++
	}
}

// Dequeue 尝试出队一个元素（单消费者）。
// ok=false 表示当前没有可消费元素（可能队列为空，或“下一槽”尚未就绪）。
//
// 并发语义：
// - 消费者先观察 ready 位（LoadUint64），位为 1 才读取 data[idx]；
// - 读取成功后递增 r，清理 data 槽位的引用（帮助 GC），并减少 size/ready；
// - ready 位不在逐元素消费路径中清理：r 单调递增使得已消费槽不会被再次检查；ready 位在分段回收（putSeg）时批量复位。
func (q *Queue[T]) Dequeue() (T, bool) {
	var zero T

	// 增加活动操作数，确保 Close 时安全回收
	q.activeOps.Add(1)
	defer q.activeOps.Add(-1)

	// 检查队列是否已关闭
	if q.closed.Load() {
		return zero, false
	}

	for {
		h := q.head
		if h == nil {
			return zero, false
		}

		// 尝试读取元素
		if h.r < uint64(len(h.data)) {
			if !h.isReadyAt(h.r) {
				return zero, false
			}
			v := h.data[h.r]
			h.data[h.r] = zero
			h.r++
			q.size.Add(-1)
			q.ready.Add(-1)
			return v, true
		}

		// 尝试推进分段
		nx := h.next.Load()
		if nx == nil {
			return zero, false
		}
		q.head = nx
		q.retireSeg(h)
	}
}

// Close 关闭队列，清理并回收队列持有的所有分段，标记队列不再被使用。
// 调用 Close 后，后续的 Enqueue 和 Dequeue 调用将直接返回 false。
// 该方法支持与 Enqueue/Dequeue 并发调用，内部会等待正在执行的操作完成后再安全回收资源。
func (q *Queue[T]) Close() {
	if !q.closed.CompareAndSwap(false, true) {
		return
	}

	// 等待所有正在执行的 Enqueue/Dequeue 操作完成，确保安全回收分段
	for q.activeOps.Load() > 0 {
		runtime.Gosched()
	}

	h := q.head
	q.head = nil
	q.tail.Store(nil)

	for h != nil {
		nx := h.next.Load()
		q.retireSeg(h)
		h = nx
	}

	q.size.Store(0)
	q.ready.Store(0)
}

// Size 返回当前队列大小（非负）。
// 该值是并发近似值：与 Enqueue/Dequeue 并发时可能存在瞬时偏差，但用于判空/限流是安全的。
func (q *Queue[T]) Size() int64 {
	return q.size.Load()
}

// Ready 返回当前准备就绪待消费的元素数量。
func (q *Queue[T]) Ready() int64 {
	return q.ready.Load()
}

// Cap 返回配置的在队上限；<=0 表示不限制。
func (q *Queue[T]) Cap() int64 {
	return q.cap
}

// tryAcquire 尝试占用一个“在队配额”。
// 当 cap<=0 时不限制，直接 +1；否则在达到上限时撤销并返回 false。
func (q *Queue[T]) tryAcquire() bool {
	if q.cap <= 0 {
		q.size.Add(1)
		return true
	}
	n := q.size.Add(1)
	if n <= q.cap {
		return true
	}
	q.size.Add(-1)
	return false
}

// getSeg 从共享池获取一个分段实例（不做字段重置）。
func (q *Queue[T]) getSeg() *segment[T] {
	s := q.segPool.Get().(*segment[T])
	return s
}

// initSeg 初始化一个新分段，并将其置为“队列持有 1 引用”的可用状态。
func (q *Queue[T]) initSeg(s *segment[T]) {
	s.dead.Store(false)
	s.refs.Store(1)
	s.next.Store(nil)
	s.w.Store(0)
	s.r = 0
}

// retireSeg 将分段标记为已从链上摘除（dead=true），并释放队列持有的 1 引用。
// 若此时 refs 归零，则立即回收；否则等待最后一个生产者释放引用时回收。
func (q *Queue[T]) retireSeg(s *segment[T]) {
	s.dead.Store(true)
	if s.refs.Add(-1) == 0 {
		q.putSeg(s)
	}
}

// releaseSeg 释放生产者持有的分段引用；当分段已处于 dead 且 refs 归零时，执行回收。
func (q *Queue[T]) releaseSeg(s *segment[T]) {
	if s.refs.Add(-1) != 0 {
		return
	}
	if !s.dead.Load() {
		return
	}
	q.putSeg(s)
}

// putSeg 将分段复位后放回池中。
// 注意：必须清空 data/ready 以释放对元素（尤其是指针）的引用，避免对象被池长期保留影响 GC；
// 同时复位计数器，确保池中对象不会泄露旧状态。
func (q *Queue[T]) putSeg(s *segment[T]) {
	s.dead.Store(false)
	s.refs.Store(0)
	s.next.Store(nil)
	n := uint64(len(s.data))
	// 尽可能清理 data 残留，释放指针引用
	if s.r < n {
		w := s.w.Load()
		if w > n {
			w = n
		}
		if w > s.r {
			clear(s.data[s.r:w])
		}
	}
	// 批量清理就绪位，保证回收后分段 ready 全部复位
	{
		w := s.w.Load()
		if w > n {
			w = n
		}
		if w > 0 {
			s.clearReadyRange(0, w)
		}
	}
	s.w.Store(0)
	s.r = 0
	q.segPool.Put(s)
}

// newSegment 分配一个新的分段对象（仅在池为空时走到该路径）。
func newSegment[T any](n int) *segment[T] {
	if n <= 0 {
		n = 256
	}
	words := (n + 63) / 64
	return &segment[T]{
		data:  make([]T, n),
		ready: make([]uint64, words),
	}
}

// acquire 尝试增加分段引用。
// dead=true 或者 refs<=0 时返回 false；否则 CAS 自增并返回 true。
func (s *segment[T]) acquire() bool {
	for {
		if s.dead.Load() {
			return false
		}
		old := s.refs.Load()
		if old <= 0 {
			return false
		}
		if s.refs.CompareAndSwap(old, old+1) {
			return true
		}
	}
}

// clearReadyRange 清除 [start, end) 区间的 ready 位（按 word 粒度）。
// 仅用于回收路径：与 putSeg 一起按需清理 r..w 的就绪标记，避免无谓全量清零。
// 要求：0 <= start < end，且 end <= len(data)。
func (s *segment[T]) clearReadyRange(start, end uint64) {
	if start >= end {
		return
	}

	n := uint64(len(s.data))
	if end > n {
		end = n
	}
	startWord := start >> 6
	endWord := (end - 1) >> 6
	for i := startWord; i <= endWord; i++ {
		s.ready[i] = 0
	}
}

// isReadyAt 返回 ready[idx] 是否为 1。
func (s *segment[T]) isReadyAt(idx uint64) bool {
	word := idx >> 6
	bit := uint64(1) << (idx & 63)
	return atomic.LoadUint64(&s.ready[word])&bit != 0
}

// setReadyAt 将 ready[idx] 置为 1。
func (s *segment[T]) setReadyAt(idx uint64) {
	word := idx >> 6
	bit := uint64(1) << (idx & 63)
	p := &s.ready[word]
	atomic.OrUint64(p, bit)
}

// clearReadyAt 将 ready[idx] 清为 0。
func (s *segment[T]) clearReadyAt(idx uint64) {
	word := idx >> 6
	bit := uint64(1) << (idx & 63)
	p := &s.ready[word]
	atomic.AndUint64(p, ^bit)
}
