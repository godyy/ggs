package systems

import (
	"github.com/godyy/ggs/internal/infra/actor/actors"
	"github.com/godyy/ggs/internal/infra/systems"
)

type serverModule struct{}

var Server = systems.RegisterSystem(&serverModule{})

func (m *serverModule) OnStart() {
}

func (m *serverModule) OnStop() {
}

func (m *serverModule) GetServerName(s *actors.Server) string {
	return s.ServerName
}
