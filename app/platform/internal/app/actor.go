package app

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggskit/infra/actor"
	pkgerrors "github.com/pkg/errors"
)

var (
	actorRegistry gactor.ActorRegistry
)

func startActor() error {
	var err error
	actorRegistry, err = actor.NewRegistry(redisClient)
	if err != nil {
		return pkgerrors.WithMessage(err, "new actor registry")
	}
	return nil
}

func ActorRegistry() gactor.ActorRegistry {
	return actorRegistry
}
