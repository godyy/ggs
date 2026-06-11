package router

import (
	"encoding/binary"
	"fmt"

	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggskit/infra/cluster"
	"github.com/godyy/ggskit/infra/noderouter"
)

// NodeSelector Agent 专用的节点选择封装，内部组合通用的 router.Selector。
// 负责：
//   - 仅接收和维护 Game 节点，并按 Category/ServerId 分组
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

// SetNodes 接收节点列表并进行分组，随后调用底层路由的 Set。
// all=true 表示全量替换所有分组；否则仅替换传入分组。
func (s *NodeSelector) SetNodes(nodes []*cluster.Node, all bool) {
	groups := make(map[string][]string)
	for _, n := range nodes {
		if !acceptNode(n) {
			continue
		}
		g := makeGroup(n.Category, n.ServerId)
		groups[g] = append(groups[g], n.GetNodeId())
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
		g := makeGroup(ev.Node.Category, ev.Node.ServerId)
		id := ev.Node.GetNodeId()
		switch ev.Type {
		case cluster.NodeEventAdd:
			updates[g] = append(updates[g], noderouter.UpdateOp{Type: noderouter.UpdateAdd, IDs: []string{id}})
		case cluster.NodeEventDel:
			updates[g] = append(updates[g], noderouter.UpdateOp{Type: noderouter.UpdateRemove, IDs: []string{id}})
		}
	}
	if len(updates) > 0 {
		s.base.Update(updates)
	}
}

// PickGame 按 serverId 选择 Game 组中的前 n 个候选节点ID；n<=1 返回单个候选。
func (s *NodeSelector) PickGame(serverId, playerId int64, n int) []string {
	group := makeGroup(consts.NodeGame, serverId)
	key := [8]byte{}
	binary.NativeEndian.PutUint64(key[:], uint64(playerId))
	return s.base.Pick(group, key[:], n)
}

func acceptNode(node *cluster.Node) bool {
	return node != nil && node.Category == consts.NodeGame
}

// makeGroup 组名生成：Game 使用 Category/ServerId，其它使用 Category。
func makeGroup(category string, serverId int64) string {
	if category == consts.NodeGame {
		return fmt.Sprintf("%s/%d", consts.NodeGame, serverId)
	}
	return category
}
