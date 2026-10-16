package actors

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggs/internal/infra/actor/lifecycle"
	"github.com/godyy/ggs/internal/infra/actor/model/player"
)

// Player 玩家Actor
type Player struct {
	actor.CActorWithModule[*player.Model] // 集成携带数据模型的Actor封装

	isLogin           bool           // 是否已登录.
	serverID          int64          // 玩家稳定所属的游戏服务器 ID.
	heartbeatTimerId  gactor.TimerId // 心跳定时器ID.
	lastHeartbeatTime int64          // 上一次心跳处理时间.
}

// NewPlayer 构造玩家Actor.
func NewPlayer(a actor.CActor) *Player {
	p := &Player{
		CActorWithModule: actor.NewCActorWithModule[*player.Model](a),
	}
	return p
}

// OnStart 启动行为.
func (p *Player) OnStart() error {
	serverID, ok, err := actor.GetActorServerID(p.ActorUID())
	if err != nil {
		return err
	}
	if !ok || serverID <= 0 {
		return errors.New("player server ownership not found")
	}
	p.serverID = serverID

	// 构造model
	p.Model = player.New(p.ActorUID().ID)

	if err := p.CActorWithModule.OnStart(p); err != nil {
		return err
	}

	return lifecycle.OnStart(p)
}

// OnStop 停机行为.
func (p *Player) OnStop(ctx gactor.ActorStopContext) error {
	// 调用生命周期回调
	lifecycle.OnStop(p)

	return p.CActorWithModule.OnStop(p, ctx)
}

// OnConnected 已连接行为.
func (p *Player) OnConnected() {
	lifecycle.OnConnected(p)
}

// OnDisconnected 已断开连接行为.
func (p *Player) OnDisconnected() {
	lifecycle.OnDisconnected(p)
}

// OnMigrateOut 持久化模型并保存需要恢复的短期会话状态.
func (p *Player) OnMigrateOut(ctx gactor.ActorMigrationContext) error {
	store := actor.GetMigrationStore()
	if store == nil {
		return errors.New("player migration store is not initialized")
	}

	// 迁移前建立 Mongo 持久化屏障，避免目标端读取旧模型。
	p.Model.SetAllDirty()
	if err := actor.SaveModel(p); err != nil {
		return err
	}
	return store.SavePlayerSnapshot(context.Background(), p.ActorUID(), ctx.SourceLeaseID, actor.PlayerMigrationSnapshot{
		ServerID:          p.serverID,
		IsLogin:           p.isLogin,
		LastHeartbeatTime: p.lastHeartbeatTime,
	})
}

// OnMigrateIn 恢复迁移前的登录与心跳状态.
func (p *Player) OnMigrateIn(ctx gactor.ActorMigrationContext) error {
	store := actor.GetMigrationStore()
	if store == nil {
		return errors.New("player migration store is not initialized")
	}
	snapshot, err := store.LoadPlayerSnapshot(context.Background(), p.ActorUID(), ctx.SourceLeaseID)
	if err != nil {
		return err
	}
	if snapshot.ServerID != p.serverID {
		return errors.New("player migration snapshot server id mismatch")
	}
	p.isLogin = snapshot.IsLogin
	p.lastHeartbeatTime = snapshot.LastHeartbeatTime
	if p.isLogin {
		p.startHeartbeatTimer()
	}
	// 快照保留到 TTL 结束，允许重复迁移控制包幂等恢复状态。
	return nil
}

func (p *Player) GetDB() string {
	return fmt.Sprintf("game_%d", p.serverID)
}

func (p *Player) SetAllDirty() {
	p.CActorWithModule.SetAllDirty(p)
}

func (p *Player) SetDirtyModules(mk ...actor.ModuleKey) {
	p.CActorWithModule.SetDirtyModules(p, mk...)
}

// ServerID 返回玩家稳定所属的游戏服务器 ID。
func (p *Player) ServerID() int64 {
	return p.serverID
}

// ID 获取玩家ID.
func (p *Player) ID() int64 {
	return p.GetActor().ActorUID().ID
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
func (p *Player) onHeartbeatTimer(args gactor.ActorTimerArgs) {
	if args.TID != p.heartbeatTimerId {
		return
	}
	logger.Get().Debugf("player %d heartbeat timer", p.ID())
	if time.Now().UnixNano()-p.lastHeartbeatTime >= int64(consts.HeartbeatTimeout) {
		p.SetLogout()
	}
}

func init() {
	registerDefine(gactor.NewCActorDefine(gactor.CActorDefineConfig{
		Name:              actor.CategoryPlayer.String(),
		Category:          actor.CategoryPlayer.ActorCategory(),
		Priority:          99,
		PriMessageBoxSize: 10,
		MessageBoxSize:    10,
		RecycleTime:       time.Minute * 30,
		BehaviorCreator: func(c gactor.CActor) gactor.CActorBehavior {
			return NewPlayer(c)
		},
	}))
}
