package systems

import (
	"fmt"

	"github.com/godyy/ggs/app/game/internal/actors"
	"github.com/godyy/ggs/app/game/internal/actors/model/player"
	"github.com/godyy/ggs/app/game/internal/consts"
	"github.com/godyy/ggs/internal/core/actor"
)

type playerModule struct{}

var Player = &playerModule{}

func init() {
	actors.RegisterLifeCycleCB(actor.CategoryPlayer, Player)
}

// OnStart Player OnStart回调.
func (m *playerModule) OnStart(p *actors.Player) error {
	return nil
}

// OnStop Player OnStop回调.
func (m *playerModule) OnStop(p *actors.Player) {

}

// InitPlayer 初始化player.
func (m *playerModule) InitPlayer(p *actors.Player) error {
	if p.Player.IsInit() {
		return nil
	}

	base := actors.GetModule[*player.BaseInfo](p, true)
	base.Name = fmt.Sprintf("player%d", p.ID())
	p.Player.Version = consts.VersionInit
	p.Player.SetDirtyAll()

	return nil
}
