package glog

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// CoreConfig 创建 Core 的配置函数.
type CoreConfig func(c *Config) zapcore.Core

// NewStdCoreConfig 创建指向标准输出的 CoreConfig.
func NewStdCoreConfig() CoreConfig {
	return func(c *Config) zapcore.Core {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		if c.Development {
			encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		}
		encoder := zapcore.NewConsoleEncoder(encoderConfig)
		return zapcore.NewTee(
			zapcore.NewCore(
				encoder,
				zapcore.Lock(os.Stdout),
				zap.LevelEnablerFunc(func(level zapcore.Level) bool {
					return level >= c.Level && level < ErrorLevel
				}),
			),
			zapcore.NewCore(
				encoder,
				zapcore.Lock(os.Stderr),
				zap.LevelEnablerFunc(func(level zapcore.Level) bool {
					return level >= c.Level && level >= ErrorLevel
				}),
			),
		)
	}
}

// FileCoreParams 日志文件输出相关 Core 配置参数.
type FileCoreParams struct {
	// Path 文件路径
	Path string

	// MaxSize 文件大小上限，用于切割日志文件.
	MaxSize int

	// MaxAge 文件保留的周期 天.
	MaxAge int

	// MaxBackups 最大备份数量.
	MaxBackups int

	// LocalTime 切割文件时是否使用本地时间.
	LocalTime bool

	// Compress 是否压缩日志.
	Compress bool
}

// NewFileCoreConfig 创建指向日志文件的 CoreConfig.
func NewFileCoreConfig(fc *FileCoreParams) CoreConfig {
	if fc.Path == "" {
		panic("log: FileCoreConfig: Path not specified")
	}

	return func(c *Config) zapcore.Core {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		if c.Development {
			encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		}
		return zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   fc.Path,
				MaxSize:    fc.MaxSize,
				MaxAge:     fc.MaxAge,
				MaxBackups: fc.MaxBackups,
				LocalTime:  fc.LocalTime,
				Compress:   fc.Compress,
			}),
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level >= c.Level
			}),
		)
	}
}
