package systems

import (
	"github.com/godyy/ggs/internal/infra/actors"
	"github.com/godyy/ggs/internal/infra/actors/models/player"
)

type itemsModule struct{}

var Items = &itemsModule{}

func (m *itemsModule) UseItem(p *actors.Player, itemId int32, num int64) (left int64, ok bool) {
	if itemId == 0 || num <= 0 {
		return 0, false
	}
	items := actors.GetModule[*player.Items](p, true)
	return items.Sub(itemId, num)
}
