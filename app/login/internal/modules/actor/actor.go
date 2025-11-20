package actor

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/core/actor"
	libredis "github.com/godyy/ggs/internal/libs/db/redis"
)

var (
	metaDriver *actor.MetaDriver
)

func Init() error {
	metaDriver = actor.NewMetaDriver(libredis.Inst())
	return nil
}

func AddMeta(meta *gactor.Meta) error {
	return metaDriver.AddMeta(meta)
}
