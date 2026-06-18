package systems

import (
	"fmt"

	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/actor/lifecycle"
	"github.com/godyy/ggs/internal/infra/actor/model/player"
)

type playerModule struct{}

var Player = &playerModule{}

func init() {
	lifecycle.RegisterCHandler[*actors.Player](actor.CategoryPlayer.ActorCategory(), Player)
}

// OnStart Player OnStart回调.
func (m *playerModule) OnStart(p *actors.Player) error {
	return nil
}

// OnStop Player OnStop回调.
func (m *playerModule) OnStop(p *actors.Player) {

}

// OnConnected Player OnConnected回调.
func (m *playerModule) OnConnected(p *actors.Player) {

}

// OnDisconnected Player OnDisconnected回调.
func (m *playerModule) OnDisconnected(p *actors.Player) {

}

// InitPlayer 初始化player.
func (m *playerModule) InitPlayer(p *actors.Player) error {
	if p.Model.IsInit() {
		return nil
	}

	base := actor.GetActorModule[*player.BaseInfo](p, true)
	base.Name = fmt.Sprintf("player%d", p.ID())
	Items.init(p)
	p.Model.Version = consts.VersionInit
	p.SetAllDirty()

	return nil
}
