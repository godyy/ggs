package app

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actor"
	pkgerrors "github.com/pkg/errors"
)

var (
	actorRegistry    *actor.Registry
	actorServerStore *actor.ServerStore
)

func startActor() error {
	var err error
	actorRegistry, err = actor.CreateRegistry(&actor.RegistryConfig{
		RedisCli: redisClient,
	})
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor registry")
	}
	actorServerStore, err = actor.CreateServerStore(&actor.ServerStoreConfig{
		RedisCli: redisClient,
	})
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
