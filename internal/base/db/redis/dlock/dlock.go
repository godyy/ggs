package dlock

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Options 分布式锁配置选项
type Options struct {
	// Expiry 锁的过期时间，默认30秒
	Expiry time.Duration

	// RetryDelay 获取锁失败时的重试间隔，默认100毫秒
	RetryDelay time.Duration
}

// DefaultOptions 返回默认配置
func DefaultOptions() *Options {
	return &Options{
		Expiry:     30 * time.Second,
		RetryDelay: 100 * time.Millisecond,
	}
}

// Lock Redis分布式锁实现
type Lock struct {
	client redis.UniversalClient
	key    string
	value  string
	opts   *Options
	locked bool
}

// New 创建一个新的分布式锁实例
// value 锁的值，用于标识锁的持有者，如果为空则自动生成UUID
func New(client redis.UniversalClient, key, value string, opts *Options) *Lock {
	// 如果 value 为空, 自动生成.
	if value == "" {
		value = GenerateUUID()
	}

	// 生成默认选项.
	if opts == nil {
		opts = DefaultOptions()
	}

	return &Lock{
		client: client,
		key:    key,
		value:  value,
		opts:   opts,
		locked: false,
	}
}

// Key 返回锁的键名
func (dl *Lock) Key() string {
	return dl.key
}

// IsLocked 检查锁是否被当前实例持有
func (dl *Lock) IsLocked() bool {
	return dl.locked
}

// Lock 获取锁，如果获取失败会阻塞直到获取成功或超时
func (dl *Lock) Lock(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := dl.TryLock(ctx); err == nil {
			return nil
		}

		// 等待一段时间后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(dl.opts.RetryDelay):
			continue
		}
	}
}

// TryLock 尝试获取锁，不会阻塞
func (dl *Lock) TryLock(ctx context.Context) error {
	result := dl.client.SetNX(ctx, dl.key, dl.value, dl.opts.Expiry)
	if err := result.Err(); err != nil {
		return err
	}

	if result.Val() {
		dl.locked = true
		return nil
	}

	return ErrLockNotObtained
}

// Unlock 释放锁
func (dl *Lock) Unlock(ctx context.Context) error {
	if !dl.locked {
		return ErrLockNotHeld
	}

	result := unlockScript.Run(ctx, dl.client, []string{dl.key}, dl.value)
	if err := result.Err(); err != nil {
		return err
	}

	if result.Val().(int64) == 1 {
		dl.locked = false
		return nil
	}

	return ErrLockNotHeld
}

// Refresh 刷新锁的过期时间
func (dl *Lock) Refresh(ctx context.Context) error {
	if !dl.locked {
		return ErrLockNotHeld
	}

	expiry := dl.opts.Expiry.Milliseconds()
	result := refreshScript.Run(ctx, dl.client, []string{dl.key}, dl.value, expiry)
	if err := result.Err(); err != nil {
		return err
	}

	if result.Val().(int64) == 1 {
		return nil
	}

	return ErrLockNotHeld
}
