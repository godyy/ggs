package c2s

import (
	"errors"
	"fmt"

	prototypes "github.com/godyy/ggs/internal/proto/types"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// DecodeMessage 解码数据包中的消息.
func DecodeMessage(p []byte) (proto.Message, error) {
	// 检查包长
	if len(p) < HeadLen {
		return nil, errors.New("packet length is too short")
	}

	// 获取协议ID
	pid := HeadGetPid(p)

	// 创建协议实例
	m, err := prototypes.C2S.Create(pid)
	if err != nil {
		return nil, err
	}

	// 解码消息
	if err = proto.Unmarshal(p[HeadLen:], m); err != nil {
		return nil, pkgerrors.WithMessagef(err, "unmarshal proto pid=%d", pid)
	}

	return m, nil
}

// EncodePacket 编码数据包.
func EncodePacket(pt int8, seq uint32, m proto.Message) (p []byte, err error) {
	pid, exists := prototypes.C2S.GetPid(m)
	if !exists {
		return nil, fmt.Errorf("pid of %s not found", proto.MessageName(m))
	}

	protoBytes, err := proto.Marshal(m)
	if err != nil {
		return nil, pkgerrors.WithMessagef(err, "marshal proto pid=%d", pid)
	}

	p = make([]byte, HeadLen+len(protoBytes))
	HeadSetPt(p, pt)
	HeadSetSeq(p, seq)
	HeadSetPid(p, pid)
	copy(p[HeadLen:], protoBytes)

	return
}

// EncodeReqPacket 编码请求数据包.
func EncodeReqPacket(seq uint32, m proto.Message) ([]byte, error) {
	return EncodePacket(PtReq, seq, m)
}

// EncodeRespPacket 编码响应数据包.
func EncodeRespPacket(seq uint32, m proto.Message) ([]byte, error) {
	return EncodePacket(PtResp, seq, m)
}

// EncodePushPacket 编码推送数据包.
func EncodePushPacket(m proto.Message) ([]byte, error) {
	return EncodePacket(PtPush, 0, m)
}

// // DecodePacket 解码数据包.
// func DecodePacket(p []byte) (h Head, m proto.Message, err error) {
// 	if len(p) <= HeadLen {
// 		return h, m, errors.New("packet length is too short")
// 	}

// 	copy(h[:], p[:HeadLen])
// 	pid := h.GetPid()

// 	m, err = prototypes.C2S.Create(pid)
// 	if err != nil {
// 		return h, m, err
// 	}

// 	if err = proto.Unmarshal(p[HeadLen:], m); err != nil {
// 		return h, m, pkgerrors.WithMessagef(err, "unmarshal proto pid=%d", pid)
// 	}

// 	return
// }
