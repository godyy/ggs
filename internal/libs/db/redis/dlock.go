package redis

import "github.com/godyy/ggs/internal/libs/db/redis/dlock"

// NewDLock 创建分布式锁.
func NewDLock(key, value string, opts *dlock.Options) *dlock.Lock {
	return dlock.New(client, key, value, opts)
}
