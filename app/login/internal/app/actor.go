package app

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggskit/infra/actor"
	pkgerrors "github.com/pkg/errors"
)

var (
	actorRegistry    gactor.ActorRegistry
	actorServerStore *actor.ServerStore
)

func startActor() error {
	var err error
	actorRegistry, err = actor.NewRegistry(redisClient)
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor registry")
	}
	actorServerStore, err = actor.NewServerStore(redisClient)
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor server store")
	}
	return nil
}

func ActorRegistry() gactor.ActorRegistry {
	return actorRegistry
}

func ActorServerStore() *actor.ServerStore {
	return actorServerStore
}
