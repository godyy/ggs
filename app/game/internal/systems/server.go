package systems

import "github.com/godyy/ggs/internal/infra/actor/actors"

type serverModule struct{}

var Server = &serverModule{}

func (m *serverModule) GetServerName(s *actors.Server) string {
	return s.ServerName
}
