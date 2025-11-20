package app

import (
	libredis "github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/modules/actor"
)

var (
	actorMetaDriver *actor.MetaDriver
)

func startActor() error {
	actorMetaDriver = actor.NewMetaDriver(libredis.Inst())
	return nil
}

func ActorMetaDriver() *actor.MetaDriver {
	return actorMetaDriver
}
