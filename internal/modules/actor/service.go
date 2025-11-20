package actor

import (
	"context"
	"fmt"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/actor"
	prototypes "github.com/godyy/ggs/internal/proto/types"
	"github.com/godyy/glog"
	"google.golang.org/protobuf/proto"
)

// ServiceConfig Actor服务配置.
type ServiceConfig struct {
	// Core 核心配置.
	Core *gactor.ServiceConfig

	// Logger 日志记录器.
	Logger glog.Logger
}

// Service 封装gactor.Service.
type Service struct {
	core *gactor.Service
}

// NewService 创建Actor服务.
func NewService(cfg *ServiceConfig) *Service {
	return &Service{
		core: gactor.NewService(cfg.Core, gactor.WithServiceLogger(cfg.Logger)),
	}
}

// Start 启动Actor服务.
func (s *Service) Start() error {
	return s.core.Start()
}

// Stop 停止Actor服务.
func (s *Service) Stop() error {
	return s.core.Stop()
}

// HandlePacket 处理节点字节数据.
func (s *Service) HandlePacket(remoteNodeId string, data []byte) error {
	return s.core.HandlePacket(remoteNodeId, data)
}

// StartActor 启动Actor.
func (s *Service) StartActor(ctx context.Context, uid gactor.ActorUID) error {
	return s.core.StartActor(ctx, uid)
}

// RPC 同步RPC调用.
func (s *Service) RPC(ctx context.Context, to gactor.ActorUID, args proto.Message) (proto.Message, error) {
	pid, ok := prototypes.S2S.GetPid(args)
	if !ok {
		return nil, fmt.Errorf("args %T not registered", args)
	}

	var (
		argsPayload  actor.S2SPayload
		replyPayload actor.S2SPayload
	)
	argsPayload.PID = pid
	argsPayload.Msg = args

	if err := s.core.RPC(ctx, to, &argsPayload, &replyPayload); err != nil {
		return nil, err
	}

	return replyPayload.Msg, nil
}

// AsyncRPC 异步RPC调用.
func (s *Service) AsyncRPC(ctx context.Context, to gactor.ActorUID, args proto.Message, callback func(reply proto.Message, err error)) error {
	pid, ok := prototypes.S2S.GetPid(args)
	if !ok {
		return fmt.Errorf("args %T not registered", args)
	}

	var (
		argsPayload actor.S2SPayload
	)
	argsPayload.PID = pid
	argsPayload.Msg = args

	if err := s.core.AsyncRPC(ctx, to, &argsPayload, func(r *gactor.RPCResp) {
		if err := r.Err(); err != nil {
			callback(nil, err)
			return
		}

		var replyPayload actor.S2SPayload
		if err := r.DecodeReply(&replyPayload); err != nil {
			callback(nil, err)
			return
		}

		callback(replyPayload.Msg, nil)
	}); err != nil {
		return err
	}

	return nil
}

// Cast 发送消息到目标actor.
func (s *Service) Cast(ctx context.Context, to gactor.ActorUID, msg proto.Message) error {
	pid, ok := prototypes.S2S.GetPid(msg)
	if !ok {
		return fmt.Errorf("msg %T not registered", msg)
	}
	payload := actor.S2SPayload{
		PID: pid,
		Msg: msg,
	}
	return s.core.Cast(ctx, to, &payload)
}
