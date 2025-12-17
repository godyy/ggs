package models

// Character 角色.
type Character struct {
	ID        int64  `bson:"id"`         // 角色ID
	AccountID int64  `bson:"account_id"` // 账号ID
	Name      string `bson:"name"`       // 角色名称
	ServerID  int64  `bson:"server_id"`  // 服务器ID
}
