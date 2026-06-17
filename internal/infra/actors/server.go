package actors

import (
	"fmt"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actors/models/server"
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

	// 加载model数据
	if exists, err := LoadModel(s); err != nil {
		return err
	} else if !exists {
		s.Model.SetDirty()
	}

	return nil
}

// OnStop 停机行为.
func (s *Server) OnStop() error {
	return nil
}
