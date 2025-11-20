package app

import (
	"runtime"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/gcluster"
	"github.com/godyy/ggs/app/game/internal/codec"
	mmongobd "github.com/godyy/ggs/app/game/internal/modules/mongobd"
	"github.com/godyy/ggs/internal/core/actor"
	"github.com/godyy/ggs/internal/core/cluster"
	"github.com/godyy/ggs/internal/core/db/mongobd"
	libmongo "github.com/godyy/ggs/internal/libs/db/mongo"
	"github.com/godyy/ggs/internal/libs/logger"
	pkgerrors "github.com/pkg/errors"
)

var appInst *app

func Start() error {
	appInst = &app{}

	// 启动mongo后台.
	if err := mmongobd.Start(mmongobd.Config{
		BDConfig: mongobd.BDConfig{
			Client:         libmongo.Inst(),
			Wokers:         runtime.NumCPU(),
			MaxWorkerOps:   1000,
			DefExecTimeout: time.Second * 5,
			Logger:         logger.GetLogger(),
		},
		OpChanSize:  10000,
		OpConsumers: 2,
	}); err != nil {
		return pkgerrors.WithMessage(err, "start mongobd")
	}

	// 初始化 cluster.
	if err := appInst.initCluster(); err != nil {
		return pkgerrors.WithMessage(err, "init cluster")
	}

	// 启动 Actor 服务.
	if err := appInst.startActor(); err != nil {
		return pkgerrors.WithMessage(err, "start actor")
	}

	// 启动 cluster.
	if err := appInst.startCluster(); err != nil {
		return pkgerrors.WithMessage(err, "start cluster")
	}

	return nil
}

func Stop() {
	// 停止 Actor 服务.
	appInst.stopActor()

	// 停止 cluster.
	appInst.stopCluster()
}

type app struct {
	// cluster.
	clusterCenter *cluster.Center
	clusterAgent  *gcluster.Agent

	// actor.
	actorMetaDriver *actor.MetaDriver
	actorService    *gactor.Service
	codec.Codec
}
