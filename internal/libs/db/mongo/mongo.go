package mongo

import (
	"context"
	"time"

	pkgerrors "github.com/pkg/errors"
	mongo "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

// Config 映射mongo客户端驱动配置.
type Config struct {
	// URI 数据库连接地址
	URI string

	// ConnectTimeout 连接超时时间
	ConnectTimeout time.Duration

	// MaxPoolSize 连接池最大连接数
	MaxPoolSize uint64

	// MinPoolSize 连接池最小连接数
	MinPoolSize uint64

	// MaxConnIdleTime 连接最大空闲时间
	MaxConnIdleTime time.Duration

	// ServerSelectionTimeout 服务器选择超时时间
	ServerSelectionTimeout time.Duration

	// HeartbeatInterval 心跳检测间隔时间
	HeartbeatInterval time.Duration

	// RetryWrites 是否启用写重试
	RetryWrites bool

	// RetryReads 是否启用读重试
	RetryReads bool

	// Direct 是否直连模式
	Direct bool

	// ReplicaSet 副本集名称
	ReplicaSet string

	// MaxConnecting 最大并发连接数
	MaxConnecting uint64
}

// client 客户端实例.
var client *mongo.Client

func Init(cfg *Config) error {
	opts := options.Client()

	// 设置连接URI
	opts.ApplyURI(cfg.URI)

	// 设置连接超时时间
	if cfg.ConnectTimeout > 0 {
		opts.SetConnectTimeout(cfg.ConnectTimeout)
	}

	// 设置连接池相关参数
	if cfg.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize > 0 {
		opts.SetMinPoolSize(cfg.MinPoolSize)
	}
	if cfg.MaxConnIdleTime > 0 {
		opts.SetMaxConnIdleTime(cfg.MaxConnIdleTime)
	}
	if cfg.MaxConnecting > 0 {
		opts.SetMaxConnecting(cfg.MaxConnecting)
	}

	// 设置服务器选择和心跳检测参数
	if cfg.ServerSelectionTimeout > 0 {
		opts.SetServerSelectionTimeout(cfg.ServerSelectionTimeout)
	}
	if cfg.HeartbeatInterval > 0 {
		opts.SetHeartbeatInterval(cfg.HeartbeatInterval)
	}

	// 设置重试策略
	opts.SetRetryWrites(cfg.RetryWrites)
	opts.SetRetryReads(cfg.RetryReads)

	// 设置连接模式
	opts.SetDirect(cfg.Direct)
	if cfg.ReplicaSet != "" {
		opts.SetReplicaSet(cfg.ReplicaSet)
	}

	// 默认选项.
	if opts.ReadConcern == nil {
		opts.SetReadConcern(readconcern.Majority())
	}
	if opts.ReadPreference == nil {
		opts.SetReadPreference(readpref.Secondary())
	}
	if opts.WriteConcern == nil {
		opts.SetWriteConcern(writeconcern.Majority())
	}

	// 创建客户端连接
	cli, err := mongo.Connect(opts)
	if err != nil {
		return pkgerrors.WithMessage(err, "conncet")
	}

	// 验证连接是否成功
	// 设置ping超时时间为连接超时时间和心跳间隔时间的总和
	if err = cli.Ping(context.Background(), nil); err != nil {
		return pkgerrors.WithMessage(err, "ping")
	}

	client = cli
	return nil
}

// Inst 返回客户端实例.
func Inst() *mongo.Client {
	return client
}
