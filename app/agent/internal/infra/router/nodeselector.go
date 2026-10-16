package router

import (
	"encoding/binary"

	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggskit/infra/cluster"
	"github.com/godyy/ggskit/infra/noderouter"
)

// NodeSelector Agent 专用的节点选择封装，内部组合通用的 router.Selector。
// 负责：
//   - 仅接收和维护全局共享的 Player worker 节点
//   - 承接 Center 的增量/全量事件并更新路由
//   - 提供面向业务的查询方法（如按 serverId 选 Game 节点）
type NodeSelector struct {
	base noderouter.Selector
}

// NewNodeSelector 创建 NodeSelector。
// base 为通用路由实现（如 RendezvousSelector）。
func NewNodeSelector(base noderouter.Selector) *NodeSelector {
	return &NodeSelector{base: base}
}

// SetNodes 接收节点列表并归入全局 worker 组，随后调用底层路由的 Set。
// all=true 表示全量替换所有分组；否则仅替换传入分组。
func (s *NodeSelector) SetNodes(nodes []*cluster.Node, all bool) {
	groups := make(map[string][]string)
	for _, n := range nodes {
		if !acceptNode(n) {
			continue
		}
		groups[consts.NodeGameWorker] = append(groups[consts.NodeGameWorker], n.GetNodeId())
	}
	s.base.Set(groups, all)
}

// UpdateEvents 接收 Center 的批量有序事件，按分组聚合为 UpdateOp 并调用底层路由的 Update。
// 要求同一分组内保持事件顺序。
func (s *NodeSelector) UpdateEvents(events []cluster.NodeEvent) {
	updates := make(map[string][]noderouter.UpdateOp)
	for _, ev := range events {
		if !acceptNode(ev.Node) {
			continue
		}
		id := ev.Node.GetNodeId()
		switch ev.Type {
		case cluster.NodeEventAdd:
			updates[consts.NodeGameWorker] = append(updates[consts.NodeGameWorker], noderouter.UpdateOp{Type: noderouter.UpdateAdd, IDs: []string{id}})
		case cluster.NodeEventDel:
			updates[consts.NodeGameWorker] = append(updates[consts.NodeGameWorker], noderouter.UpdateOp{Type: noderouter.UpdateRemove, IDs: []string{id}})
		}
	}
	if len(updates) > 0 {
		s.base.Update(updates)
	}
}

// PickWorker 按 playerId 从全局 Player worker 组中选择前 n 个候选节点 ID；n<=1 返回单个候选。
func (s *NodeSelector) PickWorker(playerId int64, n int) []string {
	key := [8]byte{}
	binary.NativeEndian.PutUint64(key[:], uint64(playerId))
	return s.base.Pick(consts.NodeGameWorker, key[:], n)
}

func acceptNode(node *cluster.Node) bool {
	return node != nil && node.Category == consts.NodeGameWorker
}
