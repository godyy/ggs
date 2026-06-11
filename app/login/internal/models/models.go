package models

import "github.com/godyy/ggs/internal/models"

// AccountInfo 账号信息
type AccountInfo struct {
	AccountID int64            // 账号ID
	User      *models.UserInfo // 用户信息
}
