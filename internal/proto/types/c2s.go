package types

import (
	"math"

	"github.com/godyy/ggs/internal/proto/pb/c2s"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

type c2sProtos struct {
	*Protos
}

func (p *c2sProtos) register(pid c2s.PID, proto proto.Message) {
	if pid > math.MaxUint16 {
		panic("register c2s proto: pid is too large")
	}

	if err := p.Protos.Register(uint16(pid), proto); err != nil {
		panic(pkgerrors.WithMessagef(err, "register c2s proto pid=%d", pid))
	}
}

var C2S = &c2sProtos{
	Protos: NewProtos(),
}
