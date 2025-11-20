package agent

import (
	"net"

	"github.com/godyy/ggs/app/agent/internal"
	"github.com/godyy/ggs/app/agent/internal/app"
	"github.com/godyy/ggs/app/internal/base/crypto"
	"github.com/godyy/ggs/app/internal/base/lifecycle"
	authjwt "github.com/godyy/ggs/internal/base/auth/jwt"
	"github.com/godyy/ggs/internal/base/consts"
	inet "github.com/godyy/ggs/internal/net"
	codecc2s "github.com/godyy/ggs/internal/proto/codec/c2s"
)

var (
	// tokenKey 令牌密钥.
	tokenKey any

	// packetReadWriter 数据包读写器.
	packetReadWriter *inet.PacketReadWriterWithCryptor
)

func init() {
	internal.StartAgent = func(conn net.Conn, sessionKey []byte, readInsideIndependentRoutine bool) error {
		agent, err := NewAgent(conn, sessionKey)
		if err != nil {
			return err
		}
		agent.Start(readInsideIndependentRoutine)
		return nil
	}

	lifecycle.RegisterBeforeStart(func() {
		initTokenKey()
		initPacketReadWriter()
	})
}

func initTokenKey() {
	pubKey, err := authjwt.LoadPubKey(app.Config().TokenKeyPath)
	if err != nil {
		loggerInst.Fatal("load token key, %v", err)
		return
	}
	tokenKey = pubKey
}

func initPacketReadWriter() {
	tmpKey := make([]byte, 16)
	tmpCrypto, _ := crypto.CreateAESCrypto(tmpKey)
	minLen := uint32(tmpCrypto.EncryptedLen(codecc2s.HeadLen))
	maxLen := uint32(128 * 1024)
	packetReadWriter = inet.NewPacketReadWriterWithCryptor(minLen, maxLen, consts.ReadWriteTimeout)
}
