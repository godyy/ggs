package mongobd

import (
	"errors"
	"reflect"
	"runtime"
	"sync"

	"github.com/godyy/ggskit/infra/mongobd"
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

// BD 封装mongobd.BD, 实现消费者模式.
type BD struct {
	core   *mongobd.BD     // BD 实例
	done   chan mongobd.Op // 接收已完成操作符的通道
	wg     sync.WaitGroup  // 后台消费者等待组
	logger glog.Logger     // 日志工具
}

// New 创建一个新的 BD 实例.
func New(cfg Config) (*BD, error) {
	if err := cfg.check(); err != nil {
		return nil, err
	}

	logger := cfg.Logger.Named("mongobd")
	bdCfg := cfg.BDConfig
	bdCfg.Logger = logger
	core, err := mongobd.NewBD(bdCfg)
	if err != nil {
		return nil, err
	}

	bd := &BD{
		core:   core,
		done:   make(chan mongobd.Op, cfg.OpChanSize),
		logger: logger,
	}

	bd.wg.Add(cfg.OpConsumers)
	for i := 0; i < cfg.OpConsumers; i++ {
		go bd.opConsumer()
	}

	return bd, nil
}

// Stop 停止BD.
// 关闭 done 通道, 等待所有后台消费者完成, 并停止 BD 实例.
func (bd *BD) Stop() {
	bd.core.Stop()
	close(bd.done)
	bd.wg.Wait()
}

// opConsumer 操作符消费者.
// 从 done 通道中接收操作符, 并执行其回调函数.
func (bd *BD) opConsumer() {
	defer bd.wg.Done()
	for op := range bd.done {
		bd.consumeOp(op)
	}
}

// consumeOp 消费操作符.
// 执行操作符的回调函数, 并处理可能的panic.
func (bd *BD) consumeOp(op mongobd.Op) {
	bd.logger.DebugFields("consume op",
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
			bd.logger.ErrorFields("invoke op callback panic",
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
func (bd *BD) Add(hashKey any, op mongobd.Op, callback func(mongobd.Op)) error {
	return bd.core.Add(hashKey, op, callback, bd.done)
}

// Exec 立即执行操作符.
func (bd *BD) Exec(hashKey any, op mongobd.Op) error {
	return bd.core.Exec(hashKey, op)
}
