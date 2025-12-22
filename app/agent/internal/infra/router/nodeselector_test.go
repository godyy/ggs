package router

import (
	"testing"

	"github.com/godyy/ggs/app/internal/base/consts"
	"github.com/godyy/ggs/internal/infra/cluster"
	"github.com/godyy/ggs/internal/infra/router"
	"github.com/stretchr/testify/assert"
)

func TestNodeSelector_SetNodes(t *testing.T) {
	// 准备基础 Selector (使用真实的 RendezvousSelector)
	baseSelector := router.NewRendezvousSelector()
	selector := NewNodeSelector(baseSelector)

	// 准备测试节点数据
	nodes := []*cluster.Node{
		{Category: consts.NodeGame, Name: "game1", ServerId: 101, Addr: "addr1"},
		{Category: consts.NodeGame, Name: "game2", ServerId: 101, Addr: "addr2"}, // 同 ServerId 副本
		{Category: consts.NodeGame, Name: "game3", ServerId: 102, Addr: "addr3"}, // 不同 ServerId
		{Category: "gate", Name: "gate1", ServerId: 0, Addr: "addr4"},            // 非 Game 节点
	}

	// 设置节点
	selector.SetNodes(nodes, true)

	// 验证 Game 节点分组 (ServerId=101)
	// 对于 ServerId=101，应该只能选到 game1 或 game2
	candidates := selector.PickGame(101, 12345, 10)
	t.Logf("PickGame(101, 12345, 10) = %v", candidates)
	expected := []string{cluster.MakeNodeID(consts.NodeGame, "game1"), cluster.MakeNodeID(consts.NodeGame, "game2")}
	assert.Contains(t, expected, candidates[0])
	for _, id := range candidates {
		assert.Contains(t, expected, id)
	}

	// 验证 Game 节点分组 (ServerId=102)
	candidates = selector.PickGame(102, 12345, 10)
	assert.Equal(t, 1, len(candidates))
	assert.Equal(t, cluster.MakeNodeID(consts.NodeGame, "game3"), candidates[0])

	// 验证非 Game 节点 (gate)
	// PickByCategory 应该能选到 gate1
	candidates = selector.PickByCategory("gate", []byte("key"), 10)
	assert.Equal(t, 1, len(candidates))
	assert.Equal(t, cluster.MakeNodeID("gate", "gate1"), candidates[0])
}

func TestNodeSelector_PickGame_FaultTolerance(t *testing.T) {
	baseSelector := router.NewRendezvousSelector()
	selector := NewNodeSelector(baseSelector)

	// 模拟 3 个副本
	nodes := []*cluster.Node{
		{Category: consts.NodeGame, Name: "n1", ServerId: 101},
		{Category: consts.NodeGame, Name: "n2", ServerId: 101},
		{Category: consts.NodeGame, Name: "n3", ServerId: 101},
	}
	selector.SetNodes(nodes, true)

	// 请求选 3 个，验证能否返回所有副本以供重试
	candidates := selector.PickGame(101, 999, 3)
	assert.Equal(t, 3, len(candidates))
	assert.Contains(t, candidates, cluster.MakeNodeID(consts.NodeGame, "n1"))
	assert.Contains(t, candidates, cluster.MakeNodeID(consts.NodeGame, "n2"))
	assert.Contains(t, candidates, cluster.MakeNodeID(consts.NodeGame, "n3"))
}

func TestMakeGroup(t *testing.T) {
	assert.Equal(t, consts.NodeGame+"/100", makeGroup(consts.NodeGame, 100))
	assert.Equal(t, "gate", makeGroup("gate", 100))
}
