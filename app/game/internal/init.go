package internal

import applifecycle "github.com/godyy/ggs/internal/base/lifecycle"

func init() {
	applifecycle.RegisterBeforeStart(beforeAppStart)
}

func beforeAppStart() {
	initActorDefineList()
}
