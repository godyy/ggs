package internal

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/actors"
	actorsdefine "github.com/godyy/ggs/internal/infra/actor/define"
)

// initActorDefineList 初始化Actor定义列表.
func initActorDefineList() {
	actorsdefine.RegisterDefine(
		// server.
		gactor.NewActorDefine(gactor.ActorDefineConfig{
			Name:           actor.CategoryServer.String(),
			Category:       actor.CategoryServer.ActorCategory(),
			Priority:       0,
			MessageBoxSize: 1,
			BehaviorCreator: func(a gactor.Actor) gactor.ActorBehavior {
				return actors.NewServer(a)
			},
		}),

		// player.
		gactor.NewCActorDefine(gactor.CActorDefineConfig{
			Name:           actor.CategoryPlayer.String(),
			Category:       actor.CategoryPlayer.ActorCategory(),
			Priority:       99,
			MessageBoxSize: 1,
			RecycleTime:    1,
			BehaviorCreator: func(c gactor.CActor) gactor.CActorBehavior {
				return actors.NewPlayer(c)
			},
		}),
	)
}
