package app

import (
	rediscli "github.com/godyy/ggs/internal/base/db/redis/cli"
	"github.com/godyy/ggs/internal/infra/actor"
)

var (
	actorMetaDriver *actor.MetaDriver
)

func startActor() error {
	actorMetaDriver = actor.NewMetaDriver(rediscli.Get())
	return nil
}

func ActorMetaDriver() *actor.MetaDriver {
	return actorMetaDriver
}
