package server

import (
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/game/internal/handlers"
)

var (
	handler = handlers.NewHandler()
)

func init() {

}

func Handle(ctx *gactor.Context) {
	handler.Handle(ctx)
}
