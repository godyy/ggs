package systems

import (
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/gdconf"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/model/player"
	pbc2s "github.com/godyy/ggs/internal/protocol/pb/c2s"
	"github.com/godyy/ggs/internal/protocol/pb/common"
)

type itemsModule struct{}

var Items = &itemsModule{}

func (m *itemsModule) init(p *actors.Player) {
	items := actor.GetActorModule[*player.Items](p, true)
	for _, item := range gdconf.Global().InitItems {
		items.Add(item.Id, int64(item.Count))
	}
}

func (m *itemsModule) UseItem(p *actors.Player, itemId int32, num int64) (left int64, ok bool) {
	if itemId == 0 || num <= 0 {
		return 0, false
	}

	items := actor.GetActorModule[*player.Items](p, true)
	left, ok = items.Sub(itemId, num)
	if ok {
		p.SetDirtyModules(items)
		item, _ := items.GetItem(itemId)
		if err := p.Sugared().PushRawMessage(&pbc2s.ItemNotify{
			Items: []*common.Item{
				{
					Id:    itemId,
					Count: item.Num,
				},
			},
		}); err != nil {
			logger.Get().Error(err)
		}
	}

	return
}
