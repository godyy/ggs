package config

import (
	"github.com/godyy/ggs/internal/libs/flags"
)

func init() {
	flags.String("auth-key-path", "", "auth key (pem) file path")
	flags.String("sign-key-path", "", "sign key (pem) file path")
}

func (c *Config) ApplyFlags() error {
	if path, ok := flags.GetValue[string]("auth-key-path"); ok && path != "" {
		c.AuthKeyPath = path
	}
	if path, ok := flags.GetValue[string]("sign-key-path"); ok && path != "" {
		c.SignKeyPath = path
	}
	return nil
}
