package internal

import applifecycle "github.com/godyy/ggs/app/internal/base/lifecycle"

func init() {
	applifecycle.RegisterBeforeStart(beforeAppStart)
}

func beforeAppStart() {
	initActorDefineList()
}
