package actors

import (
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actors/lifecycle"
	"github.com/godyy/ggs/internal/infra/actors/model/player"
	"github.com/godyy/ggskit/infra/actor"
)

// Player 玩家Actor
type Player struct {
	CActorWithModule[*player.Model] // 集成携带数据模型的Actor封装

	isLogin           bool           // 是否已登录.
	heartbeatTimerId  gactor.TimerId // 心跳定时器ID.
	lastHeartbeatTime int64          // 上一次心跳处理时间.
}

// NewPlayer 构造玩家Actor.
func NewPlayer(actor actor.CActor) *Player {
	p := &Player{
		CActorWithModule: NewCActorWithModule[*player.Model](actor),
	}
	return p
}

// OnStart 启动行为.
func (p *Player) OnStart() error {
	// 构造model
	p.Model = player.New(p.ActorUID().ID)

	if err := p.onStart(); err != nil {
		return err
	}

	return lifecycle.OnStart(p)
}

// OnStop 停机行为.
func (p *Player) OnStop() error {
	// 调用生命周期回调
	lifecycle.OnStop(p)

	return p.onStop()
}

// OnConnected 已连接行为.
func (p *Player) OnConnected() {
	lifecycle.OnConnected(p)
}

// OnDisconnected 已断开连接行为.
func (p *Player) OnDisconnected() {
	lifecycle.OnDisconnected(p)
}

// ID 获取玩家ID.
func (p *Player) ID() int64 {
	return p.CActorSugared.ActorUID().ID
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
	p.Disconnect()
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
