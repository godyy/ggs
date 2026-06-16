package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actors"
	"github.com/godyy/ggs/internal/infra/actors/models/player"
	pbcs "github.com/godyy/ggs/internal/protocol/pb/c2s"
	"github.com/godyy/ggskit/infra/actor"
)

// handleModifyName 修改玩家名称
func handleModifyName(c *gactor.Context, req *pbcs.ModifyNameReq) (*pbcs.ModifyNameResp, error) {
	p := actor.CtxActor[*actors.Player](c)
	m := actors.GetModule[*player.BaseInfo](p, true)
	m.Name = req.Name
	m.SetDirty()
	return &pbcs.ModifyNameResp{Name: req.Name}, nil
}
