package systems

import (
	"github.com/godyy/ggs/internal/infra/actors"
	"github.com/godyy/ggs/internal/infra/actors/model/player"
	"github.com/godyy/ggskit/infra/actor"
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
