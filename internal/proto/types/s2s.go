package types

import (
	"math"

	"github.com/godyy/ggs/internal/proto/pb/s2s"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type s2sProtos struct {
	*Protos
}

func (p *s2sProtos) register(pid s2s.PID, proto proto.Message) {
	if pid > math.MaxUint16 {
		panic("register s2s proto: pid is too large")
	}

	if err := p.Protos.Register(uint16(pid), proto); err != nil {
		panic(pkgerrors.WithMessagef(err, "register s2s proto pid=%d", pid))
	}
}

var S2S = &s2sProtos{
	Protos: NewProtos(),
}
