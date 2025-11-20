package codec

import "google.golang.org/protobuf/proto"

type C2SPayload struct {
	Pt  int8          // 数据包类型
	Seq uint32        // 序号
	PID uint16        // 协议ID
	Msg proto.Message // 携带的消息
}

type S2SPayload struct {
	PID uint16        // 协议ID
	Msg proto.Message // 携带的消息
}
