package router

import (
	"testing"

	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggskit/infra/cluster"
	"github.com/godyy/ggskit/infra/noderouter"
	"github.com/stretchr/testify/assert"
)

func TestNodeSelector_SetNodes(t *testing.T) {
	// 准备基础 Selector (使用真实的 RendezvousSelector)
	baseSelector := noderouter.NewRendezvousSelector()
	selector := NewNodeSelector(baseSelector)

	// 准备测试节点数据
	nodes := []*cluster.Node{
		{Category: consts.NodeGameWorker, Name: "worker1", Addr: "addr1"},
		{Category: consts.NodeGameWorker, Name: "worker2", Addr: "addr2"},
		{Category: consts.NodeGameWorker, Name: "worker3", Addr: "addr3"},
		{Category: consts.NodeGame, Name: "game101", ServerId: 101, Addr: "addr4"}, // 固定节点不承载 Player
		{Category: "gate", Name: "gate1", ServerId: 0, Addr: "addr5"},              // 非 worker 节点
	}

	// 设置节点
	selector.SetNodes(nodes, true)

	candidates := selector.PickWorker(12345, 10)
	expected := []string{
		cluster.MakeNodeID(consts.NodeGameWorker, "worker1"),
		cluster.MakeNodeID(consts.NodeGameWorker, "worker2"),
		cluster.MakeNodeID(consts.NodeGameWorker, "worker3"),
	}
	assert.Contains(t, expected, candidates[0])
	for _, id := range candidates {
		assert.Contains(t, expected, id)
	}

	assert.Len(t, candidates, 3)
}

func TestNodeSelector_PickWorker_FaultTolerance(t *testing.T) {
	baseSelector := noderouter.NewRendezvousSelector()
	selector := NewNodeSelector(baseSelector)

	// 模拟 3 个副本
	nodes := []*cluster.Node{
		{Category: consts.NodeGameWorker, Name: "n1"},
		{Category: consts.NodeGameWorker, Name: "n2"},
		{Category: consts.NodeGameWorker, Name: "n3"},
	}
	selector.SetNodes(nodes, true)

	// 请求选 3 个，验证能否返回所有副本以供重试
	candidates := selector.PickWorker(999, 3)
	assert.Equal(t, 3, len(candidates))
	assert.Contains(t, candidates, cluster.MakeNodeID(consts.NodeGameWorker, "n1"))
	assert.Contains(t, candidates, cluster.MakeNodeID(consts.NodeGameWorker, "n2"))
	assert.Contains(t, candidates, cluster.MakeNodeID(consts.NodeGameWorker, "n3"))
}

func TestNodeSelector_UpdateEventsIgnoreNonWorker(t *testing.T) {
	baseSelector := noderouter.NewRendezvousSelector()
	selector := NewNodeSelector(baseSelector)

	selector.UpdateEvents([]cluster.NodeEvent{
		{
			Type: cluster.NodeEventAdd,
			Node: &cluster.Node{Category: "gate", Name: "gate1", Addr: "addr1"},
		},
		{
			Type: cluster.NodeEventAdd,
			Node: &cluster.Node{Category: consts.NodeGameWorker, Name: "worker1", Addr: "addr2"},
		},
	})

	candidates := selector.PickWorker(12345, 10)
	assert.Equal(t, []string{cluster.MakeNodeID(consts.NodeGameWorker, "worker1")}, candidates)
}
