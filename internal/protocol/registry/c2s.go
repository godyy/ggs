package registry

import (
	"math"

	"github.com/godyy/ggs/internal/protocol/pb/c2s"
	"github.com/godyy/ggskit/base/protocol"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type c2sProtos struct {
	*protocol.Registry
}

func (p *c2sProtos) register(pid c2s.PID, proto proto.Message) {
	if pid > math.MaxUint16 {
		panic("register c2s proto: pid is too large")
	}

	if err := p.Registry.Register(uint16(pid), proto); err != nil {
		panic(pkgerrors.WithMessagef(err, "register c2s proto pid=%d", pid))
	}
}

var C2S = &c2sProtos{
	Registry: protocol.NewRegistry(),
}
