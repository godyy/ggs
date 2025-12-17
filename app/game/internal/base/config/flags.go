package config

import (
	"github.com/godyy/ggs/internal/base/config"
)

func init() {
	config.AddFlag("cluster-node-name", "", "cluster node name, cant be empty")
	config.AddFlag("cluster-port", 0, "cluster port, must > 0")
	config.AddFlag("http-port", 0, "http port, 0 means disable http server")
	config.AddFlag("enable-pprof", false, "enable pprof")
}

func (c *Config) ApplyFlags() error {
	if nodename, ok := config.GetFlagValue[string]("cluster-node-name"); ok && nodename != "" {
		c.Cluster.NodeName = nodename
	}
	if port, ok := config.GetFlagValue[int]("cluster-port"); ok && port > 0 {
		c.Cluster.Port = port
	}
	if port, ok := config.GetFlagValue[int]("http-port"); ok && port > 0 {
		c.HttpPort = port
	}
	if enable, ok := config.GetFlagValue[bool]("enable-pprof"); ok && enable {
		c.EnablePProf = enable
	}
	return nil
}
