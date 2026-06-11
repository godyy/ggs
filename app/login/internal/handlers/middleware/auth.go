package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/login/internal/app"
	"github.com/godyy/ggs/app/login/internal/base/consts"
	"github.com/godyy/ggs/app/login/internal/base/errs"
	"github.com/godyy/ggs/app/login/internal/infra/repo"
	"github.com/godyy/ggs/app/login/internal/models"
	"github.com/godyy/ggs/app/login/internal/utils/ginutils"
	"github.com/godyy/ggs/internal/base/logger"
	mongomodels "github.com/godyy/ggs/internal/infra/mongo/models"
	sharedmodels "github.com/godyy/ggs/internal/models"
	cginutils "github.com/godyy/ggs/internal/utils/ginutils"
	authjwt "github.com/godyy/ggskit/base/auth/jwt"
	"github.com/godyy/ggskit/utils/ctxutils"
	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	authPub        any
	authSecretOnce sync.Once
)

func getAuthSecret() any {
	authSecretOnce.Do(func() {
		pubKey, err := authjwt.LoadPubKey(app.Config().AuthKeyPath)
		if err != nil {
			logger.Get().Errorf("load auth secret key, %v", err)
			return
		}
		authPub = pubKey
	})
	return authPub
}

// Auth 验证并解析token
func Auth(c *gin.Context) {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()

	// 获取token
	tokenString := ""
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		// 提取 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			cginutils.AbortWithStatusError(c, http.StatusUnauthorized, errs.WithErrCodeMsg(errs.ErrCodeAuthFailed, "Authorization header format error"))
			return
		}
		tokenString = parts[1]
	} else {
		var ok bool
		tokenString, ok = c.GetQuery("token")
		if !ok {
			cginutils.AbortWithStatusError(c, http.StatusUnauthorized, errs.WithErrCodeMsg(errs.ErrCodeAuthFailed, "Authorization header or token query param missing"))
			return
		}
	}

	// 解析token并返回用户信息.
	userInfo, err := parseToken(tokenString)
	if err != nil {
		cginutils.AbortWithStatusError(c, http.StatusUnauthorized, errs.WithErrCodeErr(errs.ErrCodeAuthFailed, err))
		return
	}

	// 根据用户信息获取或创建账号
	account, err := getOrCreateAccount(ctx, userInfo)
	if err != nil {
		cginutils.AbortWithStatusError(c, http.StatusInternalServerError, err)
		return
	}

	// 将账号信息写入 gin.Context，供后续 handler 使用
	ginutils.SetAccountInfo(c, &models.AccountInfo{
		AccountID: account.ID,
		User:      userInfo,
	})

	c.Next()
}

// parseToken 解析token并返回用户信息
func parseToken(token string) (*sharedmodels.UserInfo, error) {
	// 解析token
	claims, err := authjwt.ParseToken(getAuthSecret(), token)
	if err != nil {
		return nil, err
	}

	// 验证issuer
	if !claims.VerifyIssuer(app.Env().Stage(), true) {
		return nil, errors.New("invalid issuer")
	}

	// 获取subject
	sub, ok := authjwt.GetSub(claims)
	if !ok {
		return nil, errors.New("subject not found")
	}

	// 解析subject
	userInfo := &sharedmodels.UserInfo{}
	if err := json.Unmarshal([]byte(sub), userInfo); err != nil {
		return nil, pkgerrors.WithMessage(err, "unmarshal subject")
	}

	return userInfo, nil
}

// getOrCreateAccount 根据用户信息获取或创建账号
func getOrCreateAccount(ctx context.Context, userInfo *sharedmodels.UserInfo) (*mongomodels.Account, error) {
	// 获取账号.
	account, err := repo.Account.GetAccountByUID(ctx, userInfo.UID)
	if err == nil {
		return account, err
	}

	if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}

	// 生成账号ID
	accountID, err := repo.IDGenerator.GenAccountID(ctx)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "gen account id")
	}

	// 创建账号
	account = &mongomodels.Account{
		ID:  accountID,
		UID: userInfo.UID,
	}
	if account, err = repo.Account.CreateOrGetAccount(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}
