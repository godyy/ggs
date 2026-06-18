package convert

import (
	"github.com/godyy/ggs/internal/infra/actor/model/player"
	pbcommon "github.com/godyy/ggs/internal/protocol/pb/common"
)

func Item2PB(item player.Item) *pbcommon.Item {
	return &pbcommon.Item{
		Id:    item.ID,
		Count: item.Num,
	}
}

func Items2PB(items []player.Item) []*pbcommon.Item {
	pb := make([]*pbcommon.Item, 0, len(items))
	for _, item := range items {
		pb = append(pb, Item2PB(item))
	}
	return pb
}
