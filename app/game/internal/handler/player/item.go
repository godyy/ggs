package player

import (
	"github.com/godyy/ggs/app/game/internal/systems"
	"github.com/godyy/ggs/internal/infra/actor"
	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/infra/actor/protocol/pb/common"
)

func handleUseItem(c *actor.Context, req *pbc2s.UseItemReq) (*pbc2s.UseItemResp, error) {
	if req.ItemId == 0 || req.Num <= 0 {
		return nil, actor.WithPbError(pbcommon.ErrCode_ECInvalidPacket)
	}

	left, ok := systems.Items.UseItem(c, req.ItemId, req.Num)
	if !ok {
		return nil, actor.WithPbError(pbcommon.ErrCode_ECItemNotEnough)
	}

	return &pbc2s.UseItemResp{
		ItemId:  req.ItemId,
		Num:     req.Num,
		LeftNum: left,
	}, nil
}
