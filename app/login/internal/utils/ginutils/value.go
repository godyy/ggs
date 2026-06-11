package ginutils

import (
	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/login/internal/models"
)

// GetValue 获取上下文值
func GetValue[Value any](c *gin.Context, key string) (val Value, ok bool) {
	var v any
	v, ok = c.Get(key)
	if ok {
		val, ok = v.(Value)
	}
	return
}

// SetAccountInfo 设置账号信息
func SetAccountInfo(c *gin.Context, accountInfo *models.AccountInfo) {
	c.Set("accountInfo", accountInfo)
}

// GetAccountInfo 获取账号信息
func GetAccountInfo(c *gin.Context) *models.AccountInfo {
	accountInfo, _ := GetValue[*models.AccountInfo](c, "accountInfo")
	return accountInfo
}
