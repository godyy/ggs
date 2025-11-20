package net

import (
	"github.com/godyy/ggs/app/internal/base/crypto"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/net"
	codecc2s "github.com/godyy/ggs/internal/proto/codec/c2s"
)

var (
	// PacketReadWriter 数据包读写器.
	PacketReadWriter *net.PacketReadWriterWithCryptor
)

func init() {
	initPacketReadWriter()
}

func initPacketReadWriter() {
	tmpKey := make([]byte, 16)
	tmpCrypto, _ := crypto.CreateAESCrypto(tmpKey)
	minLen := uint32(tmpCrypto.EncryptedLen(codecc2s.HeadLen))
	maxLen := uint32(128 * 1024)
	PacketReadWriter = net.NewPacketReadWriterWithCryptor(minLen, maxLen, consts.ReadWriteTimeout)
}
