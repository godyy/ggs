package player

import (
	"github.com/godyy/ggs/app/game/internal/handler"
	actorhandler "github.com/godyy/ggs/internal/infra/actor/handler"
	pbc2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/c2s"
	"google.golang.org/protobuf/proto"
)

func init() {
	initC2SHandler()
	initS2SHandler()
}

func initC2SHandler() {
	registerC2SFunc((*pbc2s.LoginCharacterReq)(nil), false, actorhandler.WrapReqFunc(handleLoginCharacter))
	registerC2SFunc((*pbc2s.HeartbeatReq)(nil), true, actorhandler.WrapReqFunc(handleHeartbeat))
	registerC2SFunc((*pbc2s.ModifyNameReq)(nil), true, actorhandler.WrapReqFunc(handleModifyName))
	registerC2SFunc((*pbc2s.UseItemReq)(nil), true, actorhandler.WrapReqFunc(handleUseItem))
}

func initS2SHandler() {
}

func registerC2SFunc(msg proto.Message, checkLogin bool, f ...actorhandler.HandlerFunc) {
	if checkLogin {
		handler.RegisterC2S(msg, append([]actorhandler.HandlerFunc{mdCheckLogin}, f...)...)
	} else {
		handler.RegisterC2S(msg, f...)
	}
}

func registerS2SFunc(msg proto.Message, f ...actorhandler.HandlerFunc) {
	handler.RegisterS2S(msg, f...)
}
