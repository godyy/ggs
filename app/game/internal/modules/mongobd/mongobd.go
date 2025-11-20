package mongobd

import (
	"errors"
	"reflect"
	"runtime"
	"sync"

	"github.com/godyy/ggs/internal/core/db/mongobd"
	"github.com/godyy/glog"
	"go.uber.org/zap"
)

// Config 配置.
type Config struct {
	mongobd.BDConfig

	// OpChanSize 操作符通道大小.
	// 操作符通道用于存储已完成的操作符, 以便后台处理.
	// 默认值为 10000.
	OpChanSize int

	// OpConsumers 操作符消费后台数量.
	// 默认值为 runtime.NumCPU().
	OpConsumers int
}

func (c *Config) check() error {
	if c.Client == nil {
		return errors.New("client is nil")
	}
	if c.Logger == nil {
		return errors.New("logger is nil")
	}
	if c.OpChanSize <= 0 {
		c.OpChanSize = 10000
	}
	if c.OpConsumers <= 0 {
		c.OpConsumers = runtime.NumCPU()
	}
	return nil
}

var (
	bd     *mongobd.BD     // BD 实例
	done   chan mongobd.Op // 接收已完成操作符的通道
	wg     sync.WaitGroup  // 后台消费者等待组
	logger glog.Logger     // 日志工具
)

// Start 启动 MongoDB 模块.
// 初始化 BD 实例, 启动后台消费者, 并返回可能的错误.
func Start(cfg Config) (err error) {
	if err = cfg.check(); err != nil {
		return err
	}

	logger = cfg.Logger.Named("mongobd")
	bdCfg := cfg.BDConfig
	bdCfg.Logger = logger
	bd, err = mongobd.NewBD(bdCfg)
	if err != nil {
		return err
	}

	done = make(chan mongobd.Op, cfg.OpChanSize)
	wg.Add(cfg.OpConsumers)
	for i := 0; i < cfg.OpConsumers; i++ {
		go opConsumer()
	}

	return nil
}

// Stop 停止 MongoDB 模块.
// 关闭 done 通道, 等待所有后台消费者完成, 并停止 BD 实例.
func Stop() {
	bd.Stop()
	close(done)
	wg.Wait()
}

// opConsumer 操作符消费者.
// 从 done 通道中接收操作符, 并执行其回调函数.
func opConsumer() {
	defer wg.Done()
	for op := range done {
		consumeOp(op)
	}
}

// consumeOp 消费操作符.
// 执行操作符的回调函数, 并处理可能的panic.
func consumeOp(op mongobd.Op) {
	logger.DebugFields("consume op",
		zap.Dict("op",
			zap.String("type", reflect.TypeOf(op).String()),
			zap.Any("value", op),
		),
	)

	callback := op.Callback()
	if callback == nil {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			logger.ErrorFields("invoke op callback panic",
				zap.Dict("op",
					zap.String("type", reflect.TypeOf(op).String()),
					zap.Any("value", op),
				),
				zap.Any("error", err),
				zap.StackSkip("stack", 1),
			)
		}
	}()

	callback(op)
}

// Add 将操作符添加到后台队列.
func Add(hashKey any, op mongobd.Op, callback func(mongobd.Op)) error {
	return bd.Add(hashKey, op, callback, done)
}

// Exec 立即执行操作符.
func Exec(hashKey any, op mongobd.Op) error {
	return bd.Exec(hashKey, op)
}
