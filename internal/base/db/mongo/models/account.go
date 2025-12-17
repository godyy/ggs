package models

// Account 账号
type Account struct {
	ID  int64  `bson:"id"`  // 账号ID
	UID string `bson:"uid"` // 用户ID
}
