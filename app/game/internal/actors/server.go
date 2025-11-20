package actors

import "github.com/godyy/gactor"

// Server 服务器Actor
type Server struct {
	gactor.Actor // 集成Actor.
}

// NewServer 构造服务器Actor.
func NewServer(actor gactor.Actor) *Server {
	return &Server{
		Actor: actor,
	}
}

// GetActor 获取 Actor.
func (s *Server) GetActor() gactor.Actor {
	return s.Actor
}

// OnStart 启动行为.
func (s *Server) OnStart() error {
	return nil
}

// OnStop 停机行为.
func (s *Server) OnStop() error {
	return nil
}
