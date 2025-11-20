package consts

import "time"

const (
	ActorSaveDelay      = 5 * time.Second
	ActorSaveRetryDelay = 1 * time.Second
	ActorCastTimeout    = 5 * time.Second
)
