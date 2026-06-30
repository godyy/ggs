package c2s

import (
	"reflect"

	"github.com/godyy/ggskit/base/protocol"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// Registry 全局 C2S 协议注册器.
var Registry = protocol.NewRegistry()

// register 向 C2S 注册表注册消息。
func register(msg proto.Message) {
	if _, err := Registry.Register(msg); err != nil {
		panic(pkgerrors.WithMessagef(err, "register c2s %s", reflect.TypeOf(msg)))
	}
}
