package registry

import (
	"math"

	"github.com/godyy/ggs/internal/protocol/pb/c2s"
	"github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/infra/actor"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// Registry 全局综合协议注册器.
var Registry = actor.NewProtoRegistry()

func registerC2S(pid c2s.PID, proto proto.Message) {
	if pid > math.MaxUint16 {
		panic("register c2s proto: pid is too large")
	}

	if err := Registry.C2S.Register(uint16(pid), proto); err != nil {
		panic(pkgerrors.WithMessagef(err, "register c2s proto pid=%d", pid))
	}
}

func registerS2S(pid s2s.PID, proto proto.Message) {
	if pid > math.MaxUint16 {
		panic("register s2s proto: pid is too large")
	}

	if err := Registry.S2S.Register(uint16(pid), proto); err != nil {
		panic(pkgerrors.WithMessagef(err, "register s2s proto pid=%d", pid))
	}
}
