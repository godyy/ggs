package registry

import (
	"math"

	"github.com/godyy/ggs/internal/protocol/pb/s2s"
	"github.com/godyy/ggskit/base/protocol"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type s2sProtos struct {
	*protocol.Registry
}

func (p *s2sProtos) register(pid s2s.PID, proto proto.Message) {
	if pid > math.MaxUint16 {
		panic("register s2s proto: pid is too large")
	}

	if err := p.Registry.Register(uint16(pid), proto); err != nil {
		panic(pkgerrors.WithMessagef(err, "register s2s proto pid=%d", pid))
	}
}

var S2S = &s2sProtos{
	Registry: protocol.NewRegistry(),
}
