package agent

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/agent/internal/app/internal"
	"github.com/godyy/ggs/app/agent/internal/log"
	"github.com/godyy/ggs/app/internal/crypto"
	inet "github.com/godyy/ggs/app/internal/net"
	"github.com/godyy/ggs/internal/core/crypto/aes"
	codecc2s "github.com/godyy/ggs/internal/proto/codec/c2s"
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// ErrStopped 代理已停止错误.
var ErrStopped = errors.New("agent stopped")

// maxPendingPackets
const maxPendingPackets = 10

// Agent 用户代理.
type Agent struct {
	app    internal.App // APP.
	conn   net.Conn     // 网络连接.
	crypto aes.Cryptor  // 密码工具.

	mtx              sync.RWMutex                // 互斥锁.
	playerId         int64                       // 角色ID.
	sessionId        uint32                      // 会话ID.
	chPendingPackets chan gactor.Buffer          // 待处理的上游数据包.
	stopFlag         int32                       // 停机标志.
	stopReason       pbc2s.DisconnectPush_Reason // 停机原因.
	chStop           chan struct{}               // 用于提供停止信号.
}

// NewAgent 创建Agent.
func NewAgent(app internal.App, conn net.Conn, secretKey []byte) (*Agent, error) {
	crypto, err := crypto.CreateAESCrypto(secretKey)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "create crypto failed")
	}

	return &Agent{
		app:              app,
		conn:             conn,
		crypto:           crypto,
		chPendingPackets: make(chan gactor.Buffer, maxPendingPackets),
		chStop:           make(chan struct{}),
	}, nil
}

// PlayerId Agent 关联的 PlayerId.
func (a *Agent) PlayerId() int64 {
	return a.playerId
}

// SessionId Agent 与 Player 建立会话使用的 SessionId.
func (a *Agent) SessionId() uint32 {
	return a.sessionId
}

// isConnected 返回是否已与 Actor 之间建立连接.
func (a *Agent) isConnected() bool {
	return a.playerId != 0 && a.sessionId != 0
}

// Start 启动Agent.
func (a *Agent) Start(readInsideIndependentRoutine bool) {
	go a.pendingPacketLoop()
	if readInsideIndependentRoutine {
		go a.readLoop()
	} else {
		a.readLoop()
	}
}

// readLoop 读取循环.
func (a *Agent) readLoop() {
read_loop:
	for {
		select {
		case <-a.chStop:
			break read_loop
		default:
			// 读取下游数据包
			p, err := inet.ReadAndDecryptPacket(a.conn, a.crypto)
			if err != nil {
				a.errorFields("[readLoop] read packet field", log.FldError(err))
				a.stop(pbc2s.DisconnectPush_SystemError)
				break read_loop
			}

			// 检查数据包类型
			pt := codecc2s.HeadGetPt(p)
			if !codecc2s.CheckPtC2S(pt) {
				a.errorFields("[readLoop] read non-req packet")
				a.stop(pbc2s.DisconnectPush_SystemError)
				break read_loop
			}

			// msg hook.
			if hook, err := a.handleHookMsg(p); err != nil {
				pid := codecc2s.HeadGetPid(p)
				a.errorFields("[readLoop] handle hook message failed", log.FldPid(pid), log.FldError(err))
				a.stop(pbc2s.DisconnectPush_SystemError)
				break read_loop
			} else if hook {
				continue
			}

			// forward.
			if !a.isConnected() {
				// 未连接
				a.errorFields("[readLoop] message not hooked and not connected", log.FldPid(codecc2s.HeadGetPid(p)))
				a.stop(pbc2s.DisconnectPush_SystemError)
				break read_loop
			}
			if err := a.app.ForwardPacket2Player(a.playerId, a.sessionId, p); err != nil {
				a.errorFields("[readLoop] forward packet to actor failed", log.FldError(err))
				a.stop(pbc2s.DisconnectPush_SystemError)
				break read_loop
			}
		}
	}
}

// pendingPacketLoop 上游数据包处理循环.
func (a *Agent) pendingPacketLoop() {
write_loop:
	for {
		select {
		case p := <-a.chPendingPackets:
			unread := p.UnreadData()

			// hook
			if hook, err := a.handleHookMsg(unread); err != nil {
				pid := codecc2s.HeadGetPid(unread)
				a.errorFields("[pendingPacketLoop] handle hook message failed", log.FldPid(pid), log.FldError(err))
				a.pendingPacketLoopStop(true, pbc2s.DisconnectPush_SystemError, true)
				break write_loop
			} else if hook {
				continue
			}

			// forward
			if err := inet.EncryptAndWritePacket(a.conn, unread, a.crypto); err != nil {
				a.errorFields("[pendingPacketLoop] write packet field", log.FldError(err))
				a.pendingPacketLoopStop(true, pbc2s.DisconnectPush_SystemError, false)
				break write_loop
			}

		case <-a.chStop:
			a.pendingPacketLoopStop(false, pbc2s.DisconnectPush_Unknown, true)
			break write_loop
		}
	}
}

// pendingPacketLoopStop 上游数据包处理循环停止逻辑.
func (a *Agent) pendingPacketLoopStop(active bool, reason pbc2s.DisconnectPush_Reason, notifyDisconnect bool) {
	if active {
		a.stop(reason)
	}

	if notifyDisconnect && a.stopReason != pbc2s.DisconnectPush_Unknown {
		a.pushDisconnect(a.stopReason)
	}

	a.conn.Close()
}

// handleHookMsg 处理钩子消息.
func (a *Agent) handleHookMsg(p []byte) (bool, error) {
	pid := codecc2s.HeadGetPid(p)
	if hook := getMsgHook(pid); hook != nil {
		msg, err := codecc2s.DecodeMessage(p)
		if err != nil {
			return false, err
		}
		hook(a, p, msg)
		return true, nil
	}
	return false, nil
}

// ErrPacketLen 数据包长度错误.
var ErrPacketLen = errors.New("packet len error")

// ErrPacketPt 数据包类型错误.
var ErrPacketPt = errors.New("packet pt error")

// ReceivePacket 接收上游数据包.
func (a *Agent) ReceivePacket(p gactor.Buffer) error {
	// 检查数据包
	unread := p.UnreadData()
	if len(unread) < codecc2s.HeadLen {
		return ErrPacketLen
	}
	pt := codecc2s.HeadGetPt(unread)
	if !codecc2s.CheckPtS2C(pt) {
		return ErrPacketPt
	}

	// 数据包入列
	select {
	case a.chPendingPackets <- p:
		return nil
	case <-a.chStop:
		return ErrStopped
	}
}

// Stop Agent 停机.
func (a *Agent) Stop(reason pbc2s.DisconnectPush_Reason) {
	a.stop(reason)
}

// stop 发起停机流程.
// 当 readLoop 和 pendingPacketLoop 都停止后, Agent 完全停机.
func (a *Agent) stop(reason pbc2s.DisconnectPush_Reason) {
	if !atomic.CompareAndSwapInt32(&a.stopFlag, 0, 1) {
		return
	}

	close(a.chStop)
	a.stopReason = reason
	if a.isConnected() {
		a.app.DelAgent(a)
		a.disconnectPlayer()
	}
}

// sendMessage 发送消息.
func (a *Agent) sendMessage(pt int8, seq uint32, m proto.Message) error {
	p, err := codecc2s.EncodePacket(pt, seq, m)
	if err != nil {
		return err
	}

	if err := inet.EncryptAndWritePacket(a.conn, p, a.crypto); err != nil {
		return err
	}

	return nil
}

// sendRespMessage 发送响应消息.
func (a *Agent) sendRespMessage(seq uint32, m proto.Message) error {
	return a.sendMessage(codecc2s.PtResp, seq, m)
}

// forwardReq2Player 转发请求到 Player.
func (a *Agent) forwardReq2Player(seq uint32, m proto.Message) error {
	p, err := codecc2s.EncodeReqPacket(seq, m)
	if err != nil {
		return err
	}
	if err := a.app.ForwardPacket2Player(a.playerId, a.sessionId, p); err != nil {
		return err
	}
	return nil
}

// disconnectPlayer 断开与 Player 之间的连接.
func (a *Agent) disconnectPlayer() {
	if err := a.app.DisconnectPlayer(a.playerId, a.sessionId); err != nil {
		a.errorFields("disconnect player failed", log.FldError(err))
	}
}

// pushDisconnect 推送断开连接消息.
func (a *Agent) pushDisconnect(reason pbc2s.DisconnectPush_Reason) {
	if err := a.sendMessage(codecc2s.PtPush, 0, &pbc2s.DisconnectPush{Reason: reason}); err != nil {
		a.errorFields("send disconnect push failed", log.FldError(err))
		return
	}
}
