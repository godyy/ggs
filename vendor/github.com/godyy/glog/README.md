基于`go.uber.org/zap`的golang日志库封装.

## 用例

### 基本用法

```go
package main

import (
    "github.com/godyy/glog"
    "go.uber.org/zap"
)

func main() {
    // 创建日志配置
    config := &glog.Config{
        Level:        glog.InfoLevel,
        EnableCaller: true,
        Development:  true,
        Cores: []glog.CoreConfig{
            // 添加标准输出核心配置
            // 将 Info 和 Warn 级别的日志输出到 stdout
            // 将 Error 及以上级别的日志输出到 stderr
            glog.NewStdCoreConfig(),

            // 添加文件输出核心配置
            glog.NewFileCoreConfig(&glog.FileCoreParams{
                Path:       "./logs/app.log",  // 日志文件路径
                MaxSize:    100,              // 单个日志文件最大大小，单位：MB
                MaxAge:     7,                // 日志文件保留天数
                MaxBackups: 10,               // 最大保留的备份文件数量
                LocalTime:  true,             // 使用本地时间
                Compress:   true,             // 压缩备份文件
            }),
        },
    }

    // 创建logger
    logger := glog.NewLogger(config)

    // 基本日志输出
    logger.Info("这是一条信息日志")
    logger.Infof("这是一条带格式的信息日志: %s", "hello")
    logger.Infoln("这是一条带换行的信息日志")

    // 结构化日志
    logger.InfoFields("结构化日志", 
        zap.String("key1", "value1"),
        zap.Int("key2", 42),
    )

    // 键值对日志
    logger.Infow("键值对日志",
        "key1", "value1",
        "key2", 42,
    )
}
```

### 核心配置说明

glog 提供了两种核心配置：

1. **标准输出核心配置 (NewStdCoreConfig)**
   - 将日志输出到标准输出（stdout/stderr）
   - Info 和 Warn 级别的日志输出到 stdout
   - Error 及以上级别的日志输出到 stderr
   - 支持彩色输出
   - 使用 ISO8601 时间格式

2. **文件输出核心配置 (NewFileCoreConfig)**
   - 将日志输出到文件
   - 支持日志文件自动切割
   - 支持日志文件压缩
   - 支持日志文件保留策略
   - 使用 JSON 格式输出

### 创建子Logger

```go
// 创建带名称的子logger
childLogger := logger.Named("child")

// 创建带字段的子logger
fieldsLogger := logger.WithFields(
    zap.String("service", "user"),
    zap.String("version", "1.0.0"),
)

// 使用子logger
childLogger.Info("这是子logger的日志")
fieldsLogger.Info("这是带固定字段的日志")
```

### 不同日志级别

```go
// Debug级别
logger.Debug("调试信息")

// Info级别
logger.Info("普通信息")

// Warn级别
logger.Warn("警告信息")

// Error级别
logger.Error("错误信息")

// Panic级别
logger.Panic("严重错误，将触发panic")

// Fatal级别
logger.Fatal("致命错误，将终止程序")
```

### 开发模式

```go
config := &glog.Config{
    Development: true,  // 启用开发模式
    // ... 其他配置
}

// 在开发模式下，DPanic会触发panic
logger.DPanic("开发模式下的严重错误")
```
