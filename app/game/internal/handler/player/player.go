package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/model/player"
	"github.com/godyy/ggs/internal/protocol/pb/c2s"
)

// handleModifyName 修改玩家名称
func handleModifyName(c *gactor.Context, req *c2s.ModifyNameReq) (*c2s.ModifyNameResp, error) {
	p := actor.CtxActor[*actors.Player](c)
	m := actor.GetActorModule[*player.BaseInfo](p, true)
	oldName := m.Name
	m.Name = req.Name
	p.SetDirtyModules(m)
	logger.Get().Debugf("player %d modify name %s to %s", p.ID(), oldName, m.Name)
	return &c2s.ModifyNameResp{Name: req.Name}, nil
}
