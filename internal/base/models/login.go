package models

// UserInfo 用户信息
type UserInfo struct {
	UID string `json:"uid"` // 用户UID
}

// TokenInfo 令牌信息
type TokenInfo struct {
	UID         string `json:"uid"`          // 用户UID
	AccountID   int64  `json:"account_id"`   // 账号ID
	CharacterID int64  `json:"character_id"` // 角色ID
	ServerID    int64  `json:"server_id"`    // 服务器ID
	// todo 其它
}
