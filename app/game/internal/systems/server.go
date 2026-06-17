package systems

import "github.com/godyy/ggs/internal/infra/actors"

type serverModule struct{}

var Server = &serverModule{}

func (m *serverModule) GetServerName(s *actors.Server) string {
	return s.ServerName
}
