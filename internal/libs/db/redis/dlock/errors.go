package dlock

import "errors"

// 分布式锁相关错误
var (
	// ErrLockNotObtained 锁获取失败
	ErrLockNotObtained = errors.New("lock not obtained")

	// ErrLockNotHeld 锁未被持有
	ErrLockNotHeld = errors.New("lock not held")
)
