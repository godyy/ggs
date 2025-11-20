package internal

import (
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/actors"
	hplayer "github.com/godyy/ggs/app/game/internal/handlers/player"
	hserver "github.com/godyy/ggs/app/game/internal/handlers/server"
	"github.com/godyy/ggs/internal/base/actor"
	actordefine "github.com/godyy/ggs/internal/base/actor/define"
)

// initActorDefineList 初始化Actor定义列表.
func initActorDefineList() {
	actordefine.RegisterDefine(
		// server.
		&gactor.ActorDefine{
			ActorDefineCommon: &gactor.ActorDefineCommon{
				Name:                       actor.Category(actor.CategoryServer).String(),
				Category:                   actor.CategoryServer.Uint16(),
				Priority:                   0,
				MessageBoxSize:             1000,
				MaxTriggeredTimerAmount:    10,
				MaxCompletedAsyncRPCAmount: 10,
				RecycleTime:                0, // 不回收
				Handler:                    hserver.Handle,
			},
			BehaviorCreator: func(a gactor.Actor) gactor.ActorBehavior {
				return actors.NewServer(a)
			},
		},

		// player.
		&gactor.CActorDefine{
			ActorDefineCommon: &gactor.ActorDefineCommon{
				Name:                       actor.Category(actor.CategoryPlayer).String(),
				Category:                   actor.CategoryPlayer.Uint16(),
				Priority:                   99,
				MessageBoxSize:             10,
				MaxTriggeredTimerAmount:    10,
				MaxCompletedAsyncRPCAmount: 10,
				RecycleTime:                time.Minute * 30,
				Handler:                    hplayer.Handle,
			},
			BehaviorCreator: func(c gactor.CActor) gactor.CActorBehavior {
				return actors.NewPlayer(c)
			},
		},
	)
}
