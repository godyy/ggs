package config

import "github.com/godyy/ggs/internal/libs/flags"

func init() {
	flags.String("cluster-node-id", "", "cluster node id")
	flags.Int("cluster-port", 0, "cluster port")
	flags.Int("port", 0, "port")
}

func (c *Config) ApplyFlags() error {
	if nodeId, ok := flags.GetValue[string]("cluster-node-id"); ok && nodeId != "" {
		c.Cluster.NodeId = nodeId
	}
	if clusterPort, ok := flags.GetValue[int]("cluster-port"); ok && clusterPort > 0 {
		c.Cluster.Port = clusterPort
	}

	return nil
}
