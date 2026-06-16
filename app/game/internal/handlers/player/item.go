package player

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/base/errs"
	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/infra/actors"
	pbcs "github.com/godyy/ggs/internal/protocol/pb/c2s"
	"github.com/godyy/ggskit/infra/actor"
)

func handleUseItem(c *gactor.Context, req *pbcs.UseItemReq) (*pbcs.UseItemResp, error) {
	if req.ItemId == 0 || req.Num <= 0 {
		return nil, errs.WithC2SPbError(pbcs.ErrCode_ECInvalidPacket)
	}

	p := actor.CtxActor[*actors.Player](c)
	left, ok := systems.Items.UseItem(p, req.ItemId, req.Num)
	if !ok {
		return nil, errs.WithC2SPbError(pbcs.ErrCode_ECItemNotEnough)
	}

	return &pbcs.UseItemResp{
		ItemId:  req.ItemId,
		Num:     req.Num,
		LeftNum: left,
	}, nil
}
