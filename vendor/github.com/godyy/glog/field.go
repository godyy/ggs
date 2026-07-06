package glog

import "go.uber.org/zap"

// LogFieldParams 转换为日志字段的参数.
type LogFieldParams struct {
	Name string // 日志字段名.
	Args []any  // 自定义参数.
}

// LogField 实现了该接口的结构可以将自己转化为日志字段.
type LogField interface {
	// ToLogField 转化为日志字段..
	ToLogField(LogFieldParams) zap.Field
}
