package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/godyy/ggs/internal/models"
	codecc2s "github.com/godyy/ggs/internal/proto/codec/c2s"
	"github.com/godyy/ggs/internal/proto/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/proto/pb/common"

	"github.com/godyy/ggs/app/agent/internal/config"
	"github.com/godyy/ggs/app/agent/internal/log"
	authjwt "github.com/godyy/ggs/internal/core/auth/jwt"
	"github.com/godyy/ggs/internal/env"
	"github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/db/redis/dlock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	loginLockExpiry     = 10 * time.Second
	loginLockRetryDelay = 500 * time.Millisecond
)

var loginLockOpts = &dlock.Options{
	Expiry:     loginLockExpiry,
	RetryDelay: loginLockRetryDelay,
}

// getLoginLock 获取登录锁.
func getLoginLock(uid string) *dlock.Lock {
	return redis.NewDLock(fmt.Sprintf("agent_login_lock:%s", uid), "", loginLockOpts)
}

// handleLoginReq 处理登录请求.
func handleLoginReq(a *Agent, p []byte, msg proto.Message) {
	var (
		playerId  int64
		sessionId uint32
		seq       = codecc2s.HeadGetSeq(p)
	)

	req := msg.(*c2s.LoginReq)

	// 解析token
	tokenInfo, errcode := parseLoginToken(a, req.Token)
	if errcode != c2s.ErrCode_ECSuccess {
		a.sendRespMessage(seq, &pbcommon.Error{Code: int32(errcode)})
		a.Stop(c2s.DisconnectPush_Unknown)
		return
	}
	playerId = tokenInfo.CharacterID

	// 获取登录锁.
	lock := getLoginLock(tokenInfo.UID)

	// 创建 context.
	ctx, cancel := context.WithTimeout(context.Background(), loginLockExpiry)
	defer cancel()

	// 登录锁加锁.
	if err := lock.Lock(ctx); err != nil {
		// 加锁超时.
		a.errorFields("login lock timeout", log.FldUid(tokenInfo.UID))
		a.Stop(c2s.DisconnectPush_LoginTimeout)
		return
	}

	// 停止当前的anothor
	if anothor := a.app.GetAgent(playerId); anothor != nil {
		a.app.DelAgent(anothor)
		anothor.Stop(c2s.DisconnectPush_AnotherLogin)
	}

	// 连接 player Actor
	sessionId = a.app.GenSessionId()
	if err := a.app.Connect2Player(playerId, sessionId); err != nil {
		a.errorFields("connect player actor failed", log.FldUid(tokenInfo.UID), log.FldPlayerId(playerId), log.FldError(err))
		a.Stop(c2s.DisconnectPush_SystemError)
		return
	}

	// 更新 agent
	a.playerId = playerId
	a.sessionId = sessionId
	a.app.AddAgent(a)

	// 编码并发送登录游戏请求.
	if err := a.forwardReq2Player(seq, &c2s.LoginCharacterReq{
		Uid:       tokenInfo.UID,
		AccountId: tokenInfo.AccountID,
	}); err != nil {
		a.errorFields("forward login game request failed", log.FldUid(tokenInfo.UID), log.FldPlayerId(playerId), log.FldError(err))
		a.app.DelAgent(a)
		a.Stop(c2s.DisconnectPush_SystemError)
		lock.Unlock(ctx)
		return
	}

	// 登录锁解锁.
	lock.Unlock(ctx)
}

// parseLoginToken 解析登录令牌.
func parseLoginToken(a *Agent, token string) (*models.TokenInfo, c2s.ErrCode) {
	// 解析token
	claims, err := authjwt.ParseToken(getTokenKey(), token)
	if err != nil {
		a.InfoFields("parse token failed", zap.String("token", token), zap.NamedError("error", err))
		return nil, c2s.ErrCode_ECInvalidToken
	}

	// 验证issuer
	if !claims.VerifyIssuer(env.All().Stage(), true) {
		a.InfoFields("token issuer error", zap.String("token", token))
		return nil, c2s.ErrCode_ECInvalidToken
	}

	// 解析subject
	ti := &models.TokenInfo{}
	sub, _ := authjwt.GetSub(claims)
	if err := json.Unmarshal([]byte(sub), ti); err != nil {
		a.InfoFields("unmarshal token subject error", zap.String("token", token), zap.NamedError("error", err))
		return nil, c2s.ErrCode_ECInvalidToken
	}

	return ti, c2s.ErrCode_ECSuccess
}

// handleLoginGameResp 处理登录游戏响应.
func handleLoginGameResp(a *Agent, p []byte, msg proto.Message) {
	seq := codecc2s.HeadGetSeq(p)
	resp := msg.(*c2s.LoginCharacterResp)
	_ = resp

	// 发送登录响应.
	if err := a.sendRespMessage(seq, &c2s.LoginResp{
		// todo
	}); err != nil {
		a.errorFields("send login response failed", log.FldError(err))
		a.Stop(c2s.DisconnectPush_SystemError)
		return
	}
}

var (
	// tokenKey 令牌密钥.
	tokenKey any
	// onceLoadTokenKey 加载令牌密钥的一次执行.
	onceLoadTokenKey sync.Once
)

// getTokenKey 获取令牌密钥.
func getTokenKey() any {
	onceLoadTokenKey.Do(func() {
		pubKey, err := authjwt.LoadPubKey(config.GetConfig().TokenKeyPath)
		if err != nil {
			loggerInst.Errorf("load token key, %v", err)
			return
		}
		tokenKey = pubKey
	})

	return tokenKey
}
