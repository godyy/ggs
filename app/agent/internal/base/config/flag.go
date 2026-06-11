package config

import (
	"github.com/godyy/ggskit/base/config"
)

func init() {
	config.AddFlag("token-key-path", "", "token key path")
	config.AddFlag("port", 0, "service port, must > 0")
	config.AddFlag("cluster-port", 0, "cluster port, must > 0")
	config.AddFlag("http-port", 0, "http port, 0 means disable http server")
	config.AddFlag("enable-pprof", false, "enable pprof")
}

func (c *Config) ApplyFlags() error {
	if tokenKeyPath, ok := config.GetFlagValue[string]("token-key-path"); ok && tokenKeyPath != "" {
		c.TokenKeyPath = tokenKeyPath
	}
	if port, ok := config.GetFlagValue[int]("port"); ok && port > 0 {
		c.Port = port
	}
	if clusterPort, ok := config.GetFlagValue[int]("cluster-port"); ok && clusterPort > 0 {
		c.Cluster.Port = clusterPort
	}
	if httpPort, ok := config.GetFlagValue[int]("http-port"); ok && httpPort > 0 {
		c.HttpPort = httpPort
	}
	if enablePProf, ok := config.GetFlagValue[bool]("enable-pprof"); ok && enablePProf {
		c.EnablePProf = enablePProf
	}
	return nil
}
