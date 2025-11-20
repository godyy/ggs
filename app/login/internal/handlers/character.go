package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/godyy/gactor"
	cmodels "github.com/godyy/ggs/internal/models"

	"github.com/godyy/ggs/app/login/internal/consts"
	"github.com/godyy/ggs/app/login/internal/errs"
	mactor "github.com/godyy/ggs/app/login/internal/modules/actor"

	"github.com/godyy/ggs/internal/core/actor"
	authjwt "github.com/godyy/ggs/internal/core/auth/jwt"
	"github.com/godyy/ggs/internal/core/db/io"
	mdb "github.com/godyy/ggs/internal/core/db/models"

	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/login/httpproto"
	"github.com/godyy/ggs/app/login/internal/config"
	"github.com/godyy/ggs/app/login/internal/utils/ginutils"
	"github.com/godyy/ggs/internal/env"
	libmongo "github.com/godyy/ggs/internal/libs/db/mongo"
	libredis "github.com/godyy/ggs/internal/libs/db/redis"
	"github.com/godyy/ggs/internal/libs/db/redis/dlock"
	"github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/ggs/internal/utils/ctxutils"
	cginutils "github.com/godyy/ggs/internal/utils/ginutils"
	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type characterHandler struct {
	onceLoadSignKey sync.Once
	signKey         any
}

func init() {
	reigsterHandler(&characterHandler{})
}

// groupPath 返回路由组路径.
func (h *characterHandler) groupPath() string {
	return "/character"
}

// setupRoutes 配置路由.
func (h *characterHandler) setupRoutes(root *gin.RouterGroup, version string) {
	group := root.Group(h.groupPath())
	{
		group.GET("/list", cginutils.WrapHandlerQueryJson(h.handleCharacterList))
		group.POST("/create", cginutils.WrapHandlerJsonJson(h.handleCharacterCreate))
		group.POST("/login", cginutils.WrapHandlerJsonJson(h.handleCharacterLogin))
	}
}

func (h *characterHandler) handleCharacterList(c *gin.Context, req *httpproto.GetCharacterListReq, resp *httpproto.GetCharacterListResp) error {
	// 获取账号信息
	accountInfo := ginutils.GetAccountInfo(c)

	// 查询角色列表
	ctx := c.Request.Context()
	characters, err := io.Character.GetAllCharactersByAccountID(ctx, libmongo.Inst(), accountInfo.AccountID)
	if err != nil {
		return errs.InernalErrorWithErr(err)
	}

	resp.CharacterList = make([]httpproto.CharacterInfo, len(characters))
	for i, character := range characters {
		resp.CharacterList[i] = httpproto.CharacterInfo{
			ID:       character.ID,
			Name:     character.Name,
			ServerID: character.ServerID,
		}
	}

	return nil
}

// handleCharacterCreate 创建角色处理器.
func (h *characterHandler) handleCharacterCreate(c *gin.Context, req *httpproto.CreateCharacterReq, resp *httpproto.CreateCharacterResp) error {
	// 检查服务器是否有效
	if err := h.checkServerAvailable(req.ServerID); err != nil {
		return err
	}

	// 获取账号信息
	accountInfo := ginutils.GetAccountInfo(c)

	// 分布式加锁
	dlock, err := h.lockAccountCharacters(accountInfo.AccountID)
	if err != nil {
		logger.GetLogger().Errorf("handler [CreateCharacter], lock failed, %v", err)
		return errs.ErrCodeInternalError
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), consts.DefaultTimeout)
		defer cancel()
		if err := dlock.Unlock(ctx); err != nil {
			logger.GetLogger().Errorf("handler [CreateCharacter], unlock failed, %v", err)
		}
	}()

	// 检查创建限制条件
	if err := h.checkCreateCharacter(c, req); err != nil {
		return err
	}

	// 刷新分布式锁
	if err := h.refreshAccountCharactersLock(dlock); err != nil {
		logger.GetLogger().Errorf("handler [CreateCharacter], refresh lock after check, %v", err)
		return errs.ErrCodeInternalError
	}

	// 创建角色
	if err := h.createCharacter(c, req, resp); err != nil {
		return err
	}

	return nil
}

// handleCharacterLogin 角色登录处理器.
func (h *characterHandler) handleCharacterLogin(c *gin.Context, req *httpproto.CharacterLoginReq, resp *httpproto.CharacterLoginResp) error {
	account := ginutils.GetAccountInfo(c)

	// 获取角色
	character, err := h.getCharacter(req.CharacterID)
	if err != nil {
		return err
	}

	// 检查角色是否有效
	if character.AccountID != account.AccountID {
		return errs.ErrCodeCharacterNotExist
	}

	// todo 检查服务器状态.

	// todo 其它检查

	// 生成登录网关令牌
	tokenInfo := &cmodels.TokenInfo{
		UID:         account.User.UID,
		AccountID:   account.AccountID,
		CharacterID: req.CharacterID,
		ServerID:    character.ServerID,
	}
	token, err := h.genLoginAgentToken(tokenInfo)
	if err != nil {
		logger.GetLogger().Errorf("handler [CharacterLogin], gen login agent token failed, %v", err)
		return errs.ErrCodeInternalError
	}
	resp.Token = token

	return nil
}

// checkServerAvailable 检查服务器是否可用.
func (h *characterHandler) checkServerAvailable(serverID int64) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	server, err := io.Server.GetServer(ctx, libmongo.Inst(), serverID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errs.ErrCodeServerUnavailable
		}
		return errs.InernalErrorWithErr(pkgerrors.WithMessage(err, "query server"))
	}

	// todo
	_ = server
	return nil
}

// lockAccountCharacters 分布式锁定账号下的角色.
func (h *characterHandler) lockAccountCharacters(accountId int64) (*dlock.Lock, error) {
	lock := libredis.NewDLock(fmt.Sprintf("account_characters_lock:%d", accountId), "", &dlock.Options{
		Expiry:     10 * time.Second,
		RetryDelay: 200 * time.Millisecond,
	})

	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	if err := lock.Lock(ctx); err != nil {
		return nil, err
	}

	return lock, nil
}

// refreshAccountCharactersLock 刷新账号角色分布式锁
func (h *characterHandler) refreshAccountCharactersLock(lock *dlock.Lock) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	if err := lock.Refresh(ctx); err != nil {
		return err
	}
	return nil
}

// checkCreateCharacter 检查创建角色条件
func (h *characterHandler) checkCreateCharacter(c *gin.Context, req *httpproto.CreateCharacterReq) error {
	accountInfo := ginutils.GetAccountInfo(c)

	// 检查角色数量
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout*2)
	defer cancel()
	total, err := io.Character.GetCharacterCountByAccountID(ctx, libmongo.Inst(), accountInfo.AccountID)
	if err != nil {
		return errs.InernalErrorWithErr(pkgerrors.WithMessage(err, "get character count of account"))
	}
	if total >= consts.MaxCharacterCountPerAccount {
		return errs.ErrCodeCharacterCountLimited
	}
	serverCount, err := io.Character.GetCharacterCountByAccounIDServerID(ctx, libmongo.Inst(), accountInfo.AccountID, req.ServerID)
	if err != nil {
		return errs.InernalErrorWithErr(pkgerrors.WithMessage(err, "get character count of account server"))
	}
	if serverCount >= consts.MaxCharacterCountPerServer {
		return errs.ErrCodeServerCharacterCountLimited
	}

	return nil
}

// createCharacter 创建角色逻辑.
func (h *characterHandler) createCharacter(c *gin.Context, req *httpproto.CreateCharacterReq, resp *httpproto.CreateCharacterResp) error {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	accountInfo := ginutils.GetAccountInfo(c)

	// 生成角色ID
	ctx, cancel = ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	characterID, err := io.IDGenerator.GenCharacterID(ctx, libmongo.Inst())
	if err != nil {
		cancel()
		return errs.InernalErrorWithErr(pkgerrors.WithMessage(err, "gen character id"))
	}
	cancel()

	// 创建角色
	ctx, cancel = ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	character := &mdb.Character{
		ID:        characterID,
		AccountID: accountInfo.AccountID,
		ServerID:  req.ServerID,
	}
	if err := io.Character.CreateCharacter(ctx, libmongo.Inst(), character); err != nil {
		return errs.InernalErrorWithErr(pkgerrors.WithMessage(err, "create character"))
	}

	// 添加角色的ActorMeta信息
	if err := mactor.AddMeta(&gactor.Meta{
		Category: actor.CategoryPlayer,
		ID:       characterID,
		Deployment: gactor.NewDeploymentFollow(gactor.ActorUID{
			Category: actor.CategoryServer,
			ID:       req.ServerID,
		}),
	}); err != nil {
		return errs.InernalErrorWithErr(pkgerrors.WithMessage(err, "add actor meta"))
	}

	resp.CharacterID = characterID
	return nil
}

// getCharacter 获取角色.
func (h *characterHandler) getCharacter(characterId int64) (*mdb.Character, error) {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	character, err := io.Character.GetCharacter(ctx, libmongo.Inst(), characterId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrCodeCharacterNotExist
		}
		return nil, errs.InernalErrorWithErr(pkgerrors.WithMessage(err, "get character"))
	}
	return character, nil
}

// checkCharacterAvailable 检查角色是否有效.
func (h *characterHandler) checkCharacterAvailable(accountId int64, characterId int64) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()
	character, err := io.Character.GetCharacter(ctx, libmongo.Inst(), characterId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return errs.ErrCodeCharacterNotExist
		}
		return errs.InernalErrorWithErr(pkgerrors.WithMessage(err, "get character"))
	}
	if character.AccountID != accountId {
		return errs.ErrCodeCharacterNotExist
	}
	return nil
}

// genLoginAgentToken 生成登录agent所需的token.
func (h *characterHandler) genLoginAgentToken(tokenInfo *cmodels.TokenInfo) (string, error) {
	sub, err := json.Marshal(tokenInfo)
	if err != nil {
		return "", err
	}

	token, err := authjwt.SignToken(h.getSignKey(), env.All().Stage(), string(sub), time.Minute*5, time.Now())
	if err != nil {
		return "", err
	}

	return token, nil
}

// getSignKey 获取签名key.
// 利用once支持并发获取.
func (h *characterHandler) getSignKey() any {
	h.onceLoadSignKey.Do(func() {
		priKey, err := authjwt.LoadPrivKey(config.GetConfig().SignKeyPath)
		if err != nil {
			logger.GetLogger().Errorf("load sign key, %v", err)
			return
		}
		h.signKey = priKey
	})

	return h.signKey
}
