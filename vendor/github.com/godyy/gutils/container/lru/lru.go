package lru

import "container/list"

// Val 缓存项约束
type Val[K comparable] interface {
	LRUKey() K
}

// LRU LRU缓存容器
type LRU[K comparable, V Val[K]] struct {
	c int                 // 缓存容量
	l *list.List          // LRU链表
	m map[K]*list.Element // 缓存映射表
}

// New 构造LRU缓存容器.
func New[K comparable, V Val[K]](capacity int) *LRU[K, V] {
	return &LRU[K, V]{
		c: capacity,
		l: list.New(),
		m: make(map[K]*list.Element),
	}
}

// Put 添加或更新缓存项.
// 如果缓存已满，会删除最近最少使用的项.
func (lru *LRU[K, V]) Put(val V) (v V, rem bool) {
	key := val.LRUKey()
	e, exist := lru.m[key]
	if exist {
		e.Value = val
		lru.l.MoveToFront(e)
		return
	}
	e = lru.l.PushFront(val)
	lru.m[key] = e
	if lru.Size() > lru.c {
		// 链表尾部始终是最近最少使用的项，容量超限时将其淘汰。
		e = lru.l.Back()
		v = e.Value.(V)
		delete(lru.m, v.LRUKey())
		lru.l.Remove(e)
		return v, true
	}
	return
}

// Get 获取缓存项.
// 如果成功获取到缓存项，会将该项移动到链表头.
func (lru *LRU[K, V]) Get(key K) (v V, exist bool) {
	var e *list.Element
	e, exist = lru.m[key]
	if !exist {
		return
	}
	v = e.Value.(V)
	lru.l.MoveToFront(e)
	return
}

// Size 获取缓存大小.
func (lru *LRU[K, V]) Size() int {
	return lru.l.Len()
}

// Clear 清空缓存.
func (lru *LRU[K, V]) Clear() {
	lru.l.Init()
	lru.m = make(map[K]*list.Element)
}
