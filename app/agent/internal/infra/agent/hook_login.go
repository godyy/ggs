package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/agent/internal"
	"github.com/godyy/ggs/app/agent/internal/app"
	"github.com/godyy/ggs/app/agent/internal/base/log"
	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/infra/actor/protocol/pb/common"
	"github.com/godyy/ggs/internal/models"
	authjwt "github.com/godyy/ggskit/base/auth/jwt"
	codecc2s "github.com/godyy/ggskit/base/codec/c2s"
	"github.com/godyy/ggskit/base/db/redis"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	loginLockExpiry     = 10 * time.Second
	loginLockRetryDelay = 500 * time.Millisecond
)

var loginLockOpts = &redis.DLockOpts{
	Expiry:     loginLockExpiry,
	RetryDelay: loginLockRetryDelay,
}

// getLoginLock 获取登录锁.
func getLoginLock(uid string) *redis.DLock {
	return redis.NewDLock(app.RedisClient(), fmt.Sprintf("agent_login_lock:%s", uid), "", loginLockOpts)
}

// handleLoginReq 处理登录请求.
func handleLoginReq(a *Agent, p []byte, msg proto.Message) {
	var (
		playerId  int64
		serverId  int64
		sessionId uint32
		seq       = codecc2s.HeadGetSeq(p)
	)

	req := msg.(*pbc2s.LoginReq)

	// 解析token
	tokenInfo, errcode := parseLoginToken(a, req.Token)
	if errcode != pbc2s.ErrCode_ECSuccess {
		a.sendRespMessage(seq, &pbcommon.Error{Code: int32(errcode)})
		a.Stop(pbc2s.DisconnectPush_Unknown)
		return
	}
	playerId = tokenInfo.CharacterID
	serverId = tokenInfo.ServerID

	// 获取登录锁.
	lock := getLoginLock(tokenInfo.UID)

	// 创建 context.
	ctx, cancel := context.WithTimeout(context.Background(), loginLockExpiry)
	defer cancel()

	// 登录锁加锁.
	if err := lock.Lock(ctx); err != nil {
		// 加锁超时.
		a.errorFields("login lock timeout", log.FldUid(tokenInfo.UID))
		a.Stop(pbc2s.DisconnectPush_LoginTimeout)
		return
	}
	defer func() {
		lock.Unlock(ctx)
	}()

	// 停止当前的anothor
	if anothor := internal.GetAgent(playerId); anothor != nil {
		internal.DelAgent(anothor)
		anothor.Stop(pbc2s.DisconnectPush_AnotherLogin)
	}

	// 获取玩家 Actor 位置信息, 并按需更新节点信息.
	playerLocation, err := getPlayerLocation(playerId)
	if err != nil && !errors.Is(err, gactor.ErrActorNotExists) {
		a.errorFields("get player location failed", log.FldUid(tokenInfo.UID), log.FldPlayerId(playerId), log.FldError(err))
		a.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}
	if errors.Is(err, gactor.ErrActorNotExists) || !checkLocation(playerLocation) {
		// 选择节点.
		nodeIds := app.NodeSelector().PickGame(serverId, playerId, 1)
		if len(nodeIds) == 0 {
			a.errorFields("pick game node failed", log.FldUid(tokenInfo.UID), log.FldServerId(serverId), log.FldPlayerId(playerId))
			a.Stop(pbc2s.DisconnectPush_SystemError)
			return
		}

		a.infoFields("pick game node", log.FldUid(tokenInfo.UID), log.FldServerId(serverId), log.FldPlayerId(playerId), log.FldNodeId(nodeIds[0]))

		// 更新节点信息.
		if err := updatePlayerLocation(playerId, nodeIds[0]); err != nil {
			a.errorFields("update player location failed", log.FldUid(tokenInfo.UID), log.FldPlayerId(playerId), log.FldError(err))
			a.Stop(pbc2s.DisconnectPush_SystemError)
			return
		}
	}

	// 连接 player Actor
	sessionId = genSessionId()
	if err := connect2Player(playerId, sessionId); err != nil {
		a.errorFields("connect player actor failed", log.FldUid(tokenInfo.UID), log.FldPlayerId(playerId), log.FldError(err))
		a.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}

	// 更新 agent
	a.playerId = playerId
	a.sessionId = sessionId
	internal.AddAgent(a)

	// 编码并发送登录游戏请求.
	if err := a.forwardReq2Player(seq, &pbc2s.LoginCharacterReq{
		Uid:       tokenInfo.UID,
		AccountId: tokenInfo.AccountID,
	}); err != nil {
		a.errorFields("forward login game request failed", log.FldUid(tokenInfo.UID), log.FldPlayerId(playerId), log.FldError(err))
		internal.DelAgent(a)
		a.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}
}

// parseLoginToken 解析登录令牌.
func parseLoginToken(a *Agent, token string) (*models.TokenInfo, pbc2s.ErrCode) {
	// 解析token
	claims, err := authjwt.ParseToken(tokenKey, token)
	if err != nil {
		a.infoFields("parse token failed", zap.String("token", token), zap.NamedError("error", err))
		return nil, pbc2s.ErrCode_ECInvalidToken
	}

	// 验证issuer
	if !claims.VerifyIssuer(app.Env().Stage(), true) {
		a.infoFields("token issuer error", zap.String("token", token))
		return nil, pbc2s.ErrCode_ECInvalidToken
	}

	// 解析subject
	ti := &models.TokenInfo{}
	sub, _ := authjwt.GetSub(claims)
	if err := json.Unmarshal([]byte(sub), ti); err != nil {
		a.infoFields("unmarshal token subject error", zap.String("token", token), zap.NamedError("error", err))
		return nil, pbc2s.ErrCode_ECInvalidToken
	}

	return ti, pbc2s.ErrCode_ECSuccess
}

// handleLoginGameResp 处理登录游戏响应.
func handleLoginGameResp(a *Agent, p []byte, msg proto.Message) {
	seq := codecc2s.HeadGetSeq(p)
	resp := msg.(*pbc2s.LoginCharacterResp)
	_ = resp

	// 发送登录响应.
	if err := a.sendRespMessage(seq, &pbc2s.LoginResp{
		// todo
	}); err != nil {
		a.errorFields("send login response failed", log.FldError(err))
		a.Stop(pbc2s.DisconnectPush_SystemError)
		return
	}
}
