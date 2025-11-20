package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/actors"
	"github.com/godyy/ggs/app/game/internal/actors/model/player"
	"github.com/godyy/ggs/app/game/internal/handlers"
	pbcs "github.com/godyy/ggs/internal/proto/pb/c2s"
)

// handleModifyName 修改玩家名称
func handleModifyName(c *gactor.Context, req *pbcs.ModifyNameReq) (*pbcs.ModifyNameResp, error) {
	p := handlers.GetActor[*actors.Player](c)
	m := actors.GetModule[*player.BaseInfo](p, true)
	m.Name = req.Name
	m.SetDirty()
	return &pbcs.ModifyNameResp{Name: req.Name}, nil
}
