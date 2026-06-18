package internal

import (
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/handler"
	_ "github.com/godyy/ggs/app/game/internal/handler/player"
	_ "github.com/godyy/ggs/app/game/internal/handler/server"
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
			MessageBoxSize: 1000,
			BehaviorCreator: func(a gactor.Actor) gactor.ActorBehavior {
				return actors.NewServer(a)
			},
		},
			gactor.WithMaxTimerAmount(10),
			gactor.WithMaxAsyncRPCAmount(10),
		),

		// player.
		gactor.NewCActorDefine(gactor.CActorDefineConfig{
			Name:           actor.CategoryPlayer.String(),
			Category:       actor.CategoryPlayer.ActorCategory(),
			Priority:       99,
			MessageBoxSize: 10,
			RecycleTime:    time.Minute * 30,
			BehaviorCreator: func(c gactor.CActor) gactor.CActorBehavior {
				return actors.NewPlayer(c)
			},
		},
			gactor.WithMaxTimerAmount(10),
			gactor.WithMaxAsyncRPCAmount(10),
		),
	)
}

// GetActorHandler 获取Actor消息处理器
func GetActorHandler() gactor.HandlerFunc {
	return handler.Handle
}
