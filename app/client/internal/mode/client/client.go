package client

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	inet "github.com/godyy/ggs/app/internal/net"

	"github.com/godyy/ggs/app/client/internal/env"
	"github.com/godyy/ggs/app/client/internal/mode"
	"github.com/godyy/ggs/app/client/internal/stream"
	icrypto "github.com/godyy/ggs/app/internal/crypto"
	"github.com/godyy/ggs/internal/consts"
	codecc2s "github.com/godyy/ggs/internal/proto/codec/c2s"
	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	pbcom "github.com/godyy/ggs/internal/proto/pb/common"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const (
	stateInit  = 0 // 初始状态,负责创建角色，获取登录角色令牌等.
	stateLogin = 1 // 负责连接网关，登录游戏角色.
	statePlay  = 2 // 游玩状态. 解析命令行输入.
)

type Client struct {
	mtx        sync.Mutex
	state      int32              // 状态
	agentToken string             // 网关令牌
	stream     *stream.Stream     // stream
	seqIncr    uint32             // seq自增键
	reqSeq     uint32             // 请求Seq
	chResp     chan proto.Message // 请求响应
}

func init() {
	// 注册模块
	mode.RegisterMode("client", func() mode.Mode {
		applyFlags()
		return &Client{
			chResp: make(chan proto.Message, 1),
		}
	})
}

// Start 启动
func (c *Client) Start() {
	go c.loop()
}

// Stop 停止
func (c *Client) Stop() {

}

// changeState 改变状态
func (c *Client) changeState(state int32) {
	c.state = state
}

// loop 主循环.
func (c *Client) loop() {
	for {
		stateLogics[c.state].run(c)
	}
}

// connectAgent 连接网关.
func (c *Client) connectAgent() error {
	// 建立连接
	conn, err := net.Dial("tcp", env.AgentAddr)
	if err != nil {
		return err
	}

	// 交换密钥
	sessionKey, err := exchangeSecretKey(conn)
	if err != nil {
		conn.Close()
		return err
	}

	// 创建加密工具
	cryptor, err := icrypto.CreateAESCrypto(sessionKey)
	if err != nil {
		conn.Close()
		return err
	}

	// 创建stream
	c.stream = stream.NewStream(conn, cryptor, c)
	return nil
}

// genSeq 生成Seq.
func (c *Client) genSeq() uint32 {
	c.seqIncr++
	return c.seqIncr
}

// sendReq 发送请求, 并等待响应.
func (c *Client) sendReq(msg proto.Message) (proto.Message, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if c.stream == nil {
		return nil, pkgerrors.New("not connected")
	}

	// 清空
	select {
	case <-c.chResp:
	default:
	}

	// 发送消息.
	seq := c.genSeq()
	atomic.StoreUint32(&c.reqSeq, seq)
	if err := c.stream.SendReq(seq, msg); err != nil {
		atomic.CompareAndSwapUint32(&c.reqSeq, seq, 0)
		return nil, err
	}

	// 等待回复.
	timer := time.NewTimer(5 * time.Second)
	select {
	case rsp := <-c.chResp:
		return rsp, nil
	case <-timer.C:
		return nil, errors.New("timeout")
	}
}

// sendReq 泛型封装
func sendReq[Req, Resp proto.Message](c *Client, req Req) (r Resp, err error) {
	var resp proto.Message
	resp, err = c.sendReq(req)
	if err != nil {
		return r, err
	}

	var ok bool
	r, ok = resp.(Resp)
	if !ok {
		if respErr, ok := resp.(*pbcom.Error); ok {
			return r, fmt.Errorf("%+v", respErr)
		}
		return r, fmt.Errorf("resp is %T", resp)
	}

	return
}

func (c *Client) tick() {
	tickHeartbeat := time.NewTicker(consts.HeartbeatInterval)
	for {
		select {
		case <-tickHeartbeat.C:
			log.Println("heartbeat.")
			if _, err := sendReq[*pbc2s.HeartbeatReq, *pbc2s.HeartbeatResp](c, &pbc2s.HeartbeatReq{}); err != nil {
				log.Fatalf("send heartbeat failed, %v", err)
			}
		}
	}
}

// OnStreamMsg 处理流消息.
func (c *Client) OnStreamMsg(msg stream.Msg) {
	switch msg.Pt {
	case codecc2s.PtResp:
		log.Printf("receive resp, seq=%d, %s{%+v}", msg.Seq, reflect.TypeOf(msg.Msg).Elem().Name(), msg.Msg)
		if atomic.CompareAndSwapUint32(&c.reqSeq, msg.Seq, 0) {
			c.chResp <- msg.Msg
		}
	case codecc2s.PtPush:
		log.Printf("receive push, %s{%+v}", reflect.TypeOf(msg.Msg).Elem().Name(), msg.Msg)
	}
}

// OnStreamClose 处理流关闭事件.
func (c *Client) OnStreamClose(err error) {
	log.Fatalf("stream closed: %v", err)
}

// exchangeSecretKey 交换密钥
func exchangeSecretKey(conn net.Conn) ([]byte, error) {
	// 生成临时secret key
	tmpKey := make([]byte, 16)
	rand.Read(tmpKey)

	// 创建加密器
	entryptor, err := icrypto.CreateRSAEncryptor()
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "create encryptor failed")
	}

	// 发送临时secret key
	if err := inet.EncryptAndWritePacket(conn, tmpKey, entryptor); err != nil {
		return nil, pkgerrors.WithMessage(err, "encrypt and write tmpKey failed")
	}

	// 接收会话密钥
	sessionKeyDecryptor, err := icrypto.CreateAESCrypto(tmpKey)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "create sessionKey decryptor failed")
	}
	sessionKey, err := inet.ReadAndDecryptPacket(conn, sessionKeyDecryptor)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "read and decrypt sessionKey failed")
	}

	return sessionKey, nil
}
