package router

import (
	"hash/fnv"
	"sync"

	grv "github.com/godyy/grendezvous"
)

// RendezvousSelector 基于 rendezvous hashing 的节点路由实现.
type RendezvousSelector struct {
	mu     sync.RWMutex               // 保护 rs 的读写锁.
	rs     map[string]*grv.Rendezvous // 每个组对应一个 rendezvous 负载均衡器.
	hasher grv.Hasher                 // 节点与 key 的哈希函数.
}

// NewRendezvousSelector 创建一个基于 FNV-1a 哈希的 RendezvousSelector.
func NewRendezvousSelector() *RendezvousSelector {
	return &RendezvousSelector{
		rs: make(map[string]*grv.Rendezvous),
		hasher: func(b []byte) uint64 {
			h := fnv.New64a()
			_, _ = h.Write(b)
			return h.Sum64()
		},
	}
}

// Set 全量设置或增量设置指定组的节点列表，原有节点集合将被替换.
func (s *RendezvousSelector) Set(groups map[string][]string, all bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if all {
		s.rs = make(map[string]*grv.Rendezvous, len(groups))
	}
	for g, ids := range groups {
		if len(ids) > 0 {
			s.rs[g] = grv.New(ids, s.hasher)
		} else {
			s.rs[g] = grv.NewEmpty(s.hasher)
		}
	}
}

// Update 批量有序更新多个分组的节点集合；调用方需保证同一分组内操作顺序正确.
func (s *RendezvousSelector) Update(updates map[string][]UpdateOp) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for g, ops := range updates {
		r := s.rs[g]
		if r == nil {
			r = grv.NewEmpty(s.hasher)
			s.rs[g] = r
		}
		for _, op := range ops {
			switch op.Type {
			case UpdateAdd:
				for _, id := range op.IDs {
					r.Add(id)
				}
			case UpdateRemove:
				for _, id := range op.IDs {
					r.Remove(id)
				}
			}
		}
	}
}

// Pick 根据 key 选择前 n 个候选节点ID；当 n<=1 时返回单个候选；未知组或无候选返回空切片.
func (s *RendezvousSelector) Pick(group string, key []byte, n int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r := s.rs[group]
	if r == nil {
		return nil
	}
	if n <= 1 {
		id := r.Lookup(key)
		if id == "" {
			return nil
		}
		return []string{id}
	}
	return r.LookupN(key, n)
}
