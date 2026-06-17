package actors

import (
	"fmt"

	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/lifecycle"
	"github.com/godyy/ggs/internal/infra/actor/model/server"
)

// Server 服务器Actor
type Server struct {
	actor.ActorWithModel[*server.Model]        // 集成携带数据模型的Actor封装
	ServerName                          string // 服务器名称
}

// NewServer 构造服务器Actor.
func NewServer(a actor.Actor) *Server {
	return &Server{
		ActorWithModel: actor.NewActorWithModel[*server.Model](a),
	}
}

// OnStart 启动行为.
func (s *Server) OnStart() error {
	s.Model = server.New(s, s.ActorUID().ID)
	s.ServerName = fmt.Sprintf("server-%d", s.ActorUID().ID)

	if err := s.ActorWithModel.OnStart(); err != nil {
		return err
	}

	return lifecycle.OnStart(s)
}

// OnStop 停机行为.
func (s *Server) OnStop() error {
	// 调用生命周期回调
	lifecycle.OnStop(s)

	return s.ActorWithModel.OnStop()
}
