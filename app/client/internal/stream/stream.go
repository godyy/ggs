package stream

import (
	"net"
	"sync/atomic"

	codecc2s "github.com/godyy/ggs/internal/proto/codec/c2s"

	inet "github.com/godyy/ggs/app/internal/net"

	"github.com/godyy/ggs/internal/core/crypto/aes"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// Msg 消息包
type Msg struct {
	Pt  int8
	Seq uint32
	Msg proto.Message
}

// Handler Stream 事件处理器.
type Handler interface {
	// OnStreamMsg 处理消息.
	OnStreamMsg(Msg)

	// OnStreamClose 处理关闭事件.
	OnStreamClose(error)
}

// Stream 代理发送网络请求和接收网络消息.
type Stream struct {
	conn    net.Conn
	cryptor aes.Cryptor
	handler Handler
	closed  int32
}

// NewStream 创建Stream.
func NewStream(conn net.Conn, crypto aes.Cryptor, handler Handler) *Stream {
	s := &Stream{
		conn:    conn,
		cryptor: crypto,
		handler: handler,
	}
	go s.readLoop()
	return s
}

// SendReq 发送请求包
func (s *Stream) SendReq(seq uint32, msg proto.Message) error {
	p, err := codecc2s.EncodeReqPacket(seq, msg)
	if err != nil {
		return err
	}
	return s.Send(p)
}

// Send 发送数据包
func (s *Stream) Send(p []byte) error {
	if s.cryptor == nil {
		return inet.WritePacket(s.conn, p)
	} else {
		return inet.EncryptAndWritePacket(s.conn, p, s.cryptor)
	}
}

// readLoop 读取循环
func (s *Stream) readLoop() {
	for {
		p, err := s.readPacket()
		if err != nil {
			s.close(err)
			return
		}

		m := Msg{
			Pt:  codecc2s.HeadGetPt(p),
			Seq: codecc2s.HeadGetSeq(p),
		}
		respMsg, err := codecc2s.DecodeMessage(p)
		if err != nil {
			s.close(pkgerrors.WithMessagef(err, "decode pt:%v seq:%v", m.Pt, m.Seq))
			return
		}
		m.Msg = respMsg

		s.handler.OnStreamMsg(m)
	}
}

// readPacket 读取数据包
func (s *Stream) readPacket() ([]byte, error) {
	if s.cryptor == nil {
		return inet.ReadPacket(s.conn)
	} else {
		return inet.ReadAndDecryptPacket(s.conn, s.cryptor)
	}
}

// Close 关闭Stream
func (s *Stream) Close() {
	s.close(nil)
}

func (s *Stream) close(err error) {
	if !atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		return
	}
	s.conn.Close()
	if err != nil {
		s.handler.OnStreamClose(err)
	}
}
