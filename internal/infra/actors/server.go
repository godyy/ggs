package actors

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actors/models/server"
	"github.com/godyy/ggskit/infra/actor"
)

// Server 服务器Actor
type Server struct {
	gactor.Actor  // 集成Actor.
	*server.Model // 聚合数据模型
	persistor     // 继承数据持久化工具
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
	s.Model = server.New(s, s.ActorUID().ID)

	// 加载model数据
	if exists, err := LoadModel(s); err != nil {
		return err
	} else if !exists {
		s.SetDirty()
	}

	return nil
}

// OnStop 停机行为.
func (s *Server) OnStop() error {
	return nil
}

func (s *Server) GetModel() actor.Model {
	return s.Model
}

func (s *Server) OnModelDirty() {
	if ok, _ := s.Model.IsDirty(); !ok {
		return
	}
	DelaySave(s, ActorSaveDelay)
}
