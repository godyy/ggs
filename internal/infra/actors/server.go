package actors

import (
	"fmt"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actors/lifecycle"
	"github.com/godyy/ggs/internal/infra/actors/model/server"
)

// Server 服务器Actor
type Server struct {
	ActorWithModel[*server.Model]        // 集成携带数据模型的Actor封装
	ServerName                    string // 服务器名称
}

// NewServer 构造服务器Actor.
func NewServer(actor gactor.Actor) *Server {
	return &Server{
		ActorWithModel: NewActorWithModel[*server.Model](actor),
	}
}

// OnStart 启动行为.
func (s *Server) OnStart() error {
	s.Model = server.New(s, s.ActorUID().ID)
	s.ServerName = fmt.Sprintf("server-%d", s.ActorUID().ID)

	if err := s.onStart(); err != nil {
		return err
	}

	return lifecycle.OnStart(s)
}

// OnStop 停机行为.
func (s *Server) OnStop() error {
	// 调用生命周期回调
	lifecycle.OnStop(s)

	return s.onStop()
}
