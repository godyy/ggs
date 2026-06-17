package systems

import (
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/model/player"
)

type itemsModule struct{}

var Items = &itemsModule{}

func (m *itemsModule) UseItem(p *actors.Player, itemId int32, num int64) (left int64, ok bool) {
	if itemId == 0 || num <= 0 {
		return 0, false
	}

	items := actor.GetActorModule[*player.Items](p, true)
	left, ok = items.Sub(itemId, num)
	if ok {
		p.SetDirtyModules(items)
	}

	return
}
