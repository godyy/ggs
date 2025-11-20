package config

import (
	"github.com/BurntSushi/toml"
	pkgerrors "github.com/pkg/errors"
)

// Flags 实现该接口, 表示在配置数据解码后, 需要应用flag选项值.
type Flags interface {
	// ApplyFlags 应用 flag 选项
	ApplyFlags() error
}

// LoadFile 读取配置文件.
func LoadFile(cfg any, file string) error {
	if _, err := toml.DecodeFile(file, cfg); err != nil {
		return pkgerrors.WithMessage(err, "decode file")
	}

	if i, ok := cfg.(Flags); ok {
		if err := i.ApplyFlags(); err != nil {
			return err
		}
	}

	return nil
}
