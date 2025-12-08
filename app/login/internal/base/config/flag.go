package config

import (
	"github.com/godyy/ggs/internal/libs/config"
)

func init() {
	config.AddFlag("auth-key-path", "", "auth key (pem) file path")
	config.AddFlag("sign-key-path", "", "sign key (pem) file path")
}

func (c *Config) ApplyFlags() error {
	if path, ok := config.GetFlagValue[string]("auth-key-path"); ok && path != "" {
		c.AuthKeyPath = path
	}
	if path, ok := config.GetFlagValue[string]("sign-key-path"); ok && path != "" {
		c.SignKeyPath = path
	}
	return nil
}
