package player

import (
	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/handler"
	pbcs "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
)

func handleUseItem(c *actor.Context, req *pbcs.UseItemReq) (*pbcs.UseItemResp, error) {
	if req.ItemId == 0 || req.Num <= 0 {
		return nil, handler.WithC2SPbError(pbcs.ErrCode_ECInvalidPacket)
	}

	p := actor.CtxActor[*actors.Player](c)
	left, ok := systems.Items.UseItem(p, req.ItemId, req.Num)
	if !ok {
		return nil, handler.WithC2SPbError(pbcs.ErrCode_ECItemNotEnough)
	}

	return &pbcs.UseItemResp{
		ItemId:  req.ItemId,
		Num:     req.Num,
		LeftNum: left,
	}, nil
}
