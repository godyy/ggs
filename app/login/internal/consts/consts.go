package consts

import "time"

const (
	DefaultTimeout = time.Second * 5 // 默认超时时间，单位秒
)

const (
	MaxCharacterCountPerAccount = 2 // 每账号最大角色数量
	MaxCharacterCountPerServer  = 2 // 每服务器最大角色数量
)
