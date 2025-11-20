package redis

import (
	"context"
	"errors"
	"time"

	pkgerrors "github.com/pkg/errors"
	redis "github.com/redis/go-redis/v9"
)

// Config 映射 redis.Client 配置.
type Config struct {
	// Addrs 实例地址.
	Addrs []string

	// UserName 用户名.
	Username string

	// Password 密码.
	Password string

	// DB 数据库编号.
	DB int

	// DialTimeout 连接超时.
	DialTimeout time.Duration

	// ReadWriteTimeout 连接套接字读写超时.
	ReadWriteTimeout time.Duration

	// PoolSize 连接池大小.
	PoolSize int

	// PoolTimeout 等待连接池可用连接超时.
	PoolTimeout time.Duration

	// MinIdleConns 最小空闲连接数.
	MinIdleConns int

	// MaxIdleConns 最大空闲连接数.
	MaxIdleConns int

	// MaxActiveConns 连接池可以分配的最大连接数.
	MaxActiveConns int

	// ConnMaxIdleTime 空闲连接最大空闲时间.
	ConnMaxIdleTime time.Duration
}

func (cfg *Config) unisersal() *redis.UniversalOptions {
	return &redis.UniversalOptions{
		Addrs:           cfg.Addrs,
		Username:        cfg.Username,
		Password:        cfg.Password,
		DB:              cfg.DB,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadWriteTimeout,
		WriteTimeout:    cfg.ReadWriteTimeout,
		PoolSize:        cfg.PoolSize,
		PoolTimeout:     cfg.PoolTimeout,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		MaxActiveConns:  cfg.MaxActiveConns,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
	}
}

// client 客户端实例.
var client redis.UniversalClient

// Init 初始化.
func Init(cfg *Config) error {
	if client != nil {
		return errors.New("initialized")
	}

	cli := redis.NewUniversalClient(cfg.unisersal())
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout+cfg.ReadWriteTimeout)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		return pkgerrors.WithMessage(err, "ping")
	}

	client = cli
	return nil
}

// Inst 返回客户端实例.
func Inst() redis.UniversalClient {
	return client
}
