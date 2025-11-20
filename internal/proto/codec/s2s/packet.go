package s2s

import (
	"encoding/binary"
	"fmt"

	prototypes "github.com/godyy/ggs/internal/proto/types"

	"google.golang.org/protobuf/proto"
)

// EncodePayload 编码负载数据.
func EncodePayload(pid uint16, msg proto.Message) ([]byte, error) {
	if err := prototypes.S2S.Check(pid, msg); err != nil {
		return nil, err
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("marshal msg of pid %d failed, %v", pid, err)
	}
	payload := make([]byte, 2+len(msgBytes))
	binary.BigEndian.PutUint16(payload, uint16(pid))
	copy(payload[2:], msgBytes)
	return payload, nil
}

// DecodePayload 解码负载数据.
func DecodePayload(p []byte) (uint16, proto.Message, error) {
	if len(p) < 2 {
		return 0, nil, fmt.Errorf("payload len %d must > %d", len(p), 2)
	}
	pid := binary.BigEndian.Uint16(p)
	msg, err := prototypes.S2S.Create(pid)
	if err != nil {
		return 0, nil, fmt.Errorf("msg of pid %d not registered, %v", pid, err)
	}
	if err := proto.Unmarshal(p[2:], msg); err != nil {
		return 0, nil, fmt.Errorf("unmarshal msg of pid %d failed, %v", pid, err)
	}
	return pid, msg, nil
}
