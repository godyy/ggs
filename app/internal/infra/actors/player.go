package actors

import (
	"context"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/internal/base/consts"
	"github.com/godyy/ggs/app/internal/infra/actors/lifecycle"
	"github.com/godyy/ggs/app/internal/infra/actors/models/player"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
)

// Player 玩家Actor
type Player struct {
	gactor.CActor // 集成Actor
	*player.Model // model
	persistor     // 集成持久化辅助结构

	isLogin           bool           // 是否已登录.
	heartbeatTimerId  gactor.TimerId // 心跳定时器ID.
	lastHeartbeatTime int64          // 上一次心跳处理时间.
}

// NewPlayer 构造玩家Actor.
func NewPlayer(actor gactor.CActor) *Player {
	p := &Player{
		CActor: actor,
	}
	return p
}

// GetActor 获取 Actor.
func (p *Player) GetActor() gactor.Actor {
	return p.CActor
}

// OnStart 启动行为.
func (p *Player) OnStart() error {
	// 构造model
	p.Model = player.New(p, p.ActorUID().ID)

	// 加载model数据
	if err := LoadModel(p); err != nil {
		return err
	}

	// 调用生命周期回调的
	if err := lifecycle.OnStart(p); err != nil {
		return err
	}

	return nil
}

// OnStop 停机行为.
func (p *Player) OnStop() error {
	// 调用生命周期回调
	lifecycle.OnStop(p)

	// 持久化脏数据.
	if ok, _ := p.Model.IsDirty(); ok {
		if err := SaveModel(p); err != nil {
			// todo 日志
		}
	}

	if p.Model != nil {
		p.Model.Release()
		p.Model = nil
	}
	return nil
}

// GetCActor 获取 CActor.
func (p *Player) GetCActor() gactor.CActor {
	return p.CActor
}

// OnConnected 已连接行为.
func (p *Player) OnConnected() {
	lifecycle.OnConnected(p)
}

// OnDisconnected 已断开连接行为.
func (p *Player) OnDisconnected() {
	lifecycle.OnDisconnected(p)
}

func (p *Player) GetModel() actor.Model {
	return p.Model
}

func (p *Player) OnModelDirty() {
	if ok, _ := p.Model.IsDirty(); !ok {
		return
	}
	DelaySave(p, ActorSaveDelay)
}

func (p *Player) GetModuleContainer() actor.ModuleContainer {
	return p.Model
}

// ID 获取玩家ID.
func (p *Player) ID() int64 {
	return p.CActor.ActorUID().ID
}

// IsLogin 返回是否已登录.
func (p *Player) IsLogin() bool {
	return p.isLogin
}

// SetLogin 设置已登录.
func (p *Player) SetLogin() {
	p.isLogin = true
	p.startHeartbeatTimer()
}

// SetLogout 设置已登出.
func (p *Player) SetLogout() {
	p.isLogin = false
	p.stopHeartbeatTimer()
	p.Disconnect(context.Background())
}

// Heartbeat 心跳逻辑, 收到心跳包时调用.
func (p *Player) Heartbeat() {
	p.lastHeartbeatTime = time.Now().UnixNano()
}

// startHeartbeatTimer 启动心跳定时器.
func (p *Player) startHeartbeatTimer() {
	if p.heartbeatTimerId != gactor.TimerIdNone {
		p.StopTimer(p.heartbeatTimerId)
	}
	p.heartbeatTimerId = p.StartTimer(consts.HeartbeatTimeout, false, nil, p.onHeartbeatTimer)
}

// stopHeartbeatTimer 停止心跳定时器.
func (p *Player) stopHeartbeatTimer() {
	if p.heartbeatTimerId != gactor.TimerIdNone {
		p.StopTimer(p.heartbeatTimerId)
		p.heartbeatTimerId = gactor.TimerIdNone
	}
}

// onHeartbeatTimer 心跳定时器回调.
func (p *Player) onHeartbeatTimer(args *gactor.ActorTimerArgs) {
	if args.TID != p.heartbeatTimerId {
		return
	}
	logger.Get().Debugf("player %d heartbeat timer", p.ID())
	if time.Now().UnixNano()-p.lastHeartbeatTime >= int64(consts.HeartbeatTimeout) {
		p.SetLogout()
	}
}
