package consts

import "time"

const (
	ReadWriteTimeout = 60 * time.Second

	DefaultTimeout = 5 * time.Second

	HeartbeatInterval = 30 * time.Second

	HeartbeatTimeout = 45 * time.Second
)
