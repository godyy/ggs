package net

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/godyy/ggs/internal/base/crypto"
	pkgerrors "github.com/pkg/errors"
)

// ReadPacket 读取数据包.
func ReadPacket(conn net.Conn, timeout time.Duration) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(timeout))

	// 读取长度
	var l [4]byte
	if _, err := io.ReadFull(conn, l[:]); err != nil {
		return nil, pkgerrors.WithMessage(err, "read length")
	}
	n := binary.BigEndian.Uint32(l[:])

	// 读取负载
	p := make([]byte, n)
	if _, err := io.ReadFull(conn, p); err != nil {
		return nil, pkgerrors.WithMessage(err, "read payload")
	}

	return p, nil
}

// WritePacket 写入数据包.
func WritePacket(conn net.Conn, p []byte, timeout time.Duration) error {
	conn.SetWriteDeadline(time.Now().Add(timeout))

	// 写入长度
	var l [4]byte
	binary.BigEndian.PutUint32(l[:], uint32(len(p)))
	if err := write(conn, l[:]); err != nil {
		return pkgerrors.WithMessage(err, "write length")
	}

	// 写入负载
	if err := write(conn, p); err != nil {
		return pkgerrors.WithMessage(err, "write payload")
	}

	return nil
}

// write 将数据完整的写入
func write(w io.Writer, p []byte) error {
	total := uint32(len(p))
	written := uint32(0)
	for written < total {
		n, err := w.Write(p[written:])
		if err != nil {
			return err
		}
		written += uint32(n)
	}
	return nil
}

// PacketReadWriter 数据包读写器
type PacketReadWriter struct {
	minLen  uint32        // 最小数据包长度
	maxLen  uint32        // 最大数据包长度
	timeout time.Duration // 超时
}

// NewPacketReadWriter 创建一个新的数据包读写器
func NewPacketReadWriter(minLen, maxLen uint32, timeout time.Duration) *PacketReadWriter {
	if minLen > maxLen {
		panic("minLen must be less than or equal to maxLen")
	}
	return &PacketReadWriter{
		minLen:  minLen,
		maxLen:  maxLen,
		timeout: timeout,
	}
}

// ReadPacket 读取数据包
func (prw *PacketReadWriter) ReadPacket(conn net.Conn) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(prw.timeout))

	var l [4]byte
	if _, err := io.ReadFull(conn, l[:]); err != nil {
		return nil, pkgerrors.WithMessage(err, "read length")
	}

	len := binary.BigEndian.Uint32(l[:])
	if len < prw.minLen || len > prw.maxLen {
		return nil, fmt.Errorf("invalid packet len %d", len)
	}

	p := make([]byte, len)
	if _, err := io.ReadFull(conn, p); err != nil {
		return nil, pkgerrors.WithMessage(err, "read payload")
	}

	return p, nil
}

// WritePacket 写入数据包
func (prw *PacketReadWriter) WritePacket(conn net.Conn, p []byte) error {
	conn.SetWriteDeadline(time.Now().Add(prw.timeout))

	len := len(p)
	if len < int(prw.minLen) || len > int(prw.maxLen) {
		return fmt.Errorf("invalid packet len %d", len)
	}

	// 写入数据包长度
	var l [4]byte
	binary.BigEndian.PutUint32(l[:], uint32(len))
	if err := write(conn, l[:]); err != nil {
		return pkgerrors.WithMessage(err, "write length")
	}

	// 写入负载数据
	if err := write(conn, p); err != nil {
		return pkgerrors.WithMessage(err, "write payload")
	}

	return nil
}

// PacketReadWriterWithCryptor 带加密的数据包读写器
type PacketReadWriterWithCryptor struct {
	*PacketReadWriter
}

// NewPacketReadWriterWithCryptor 创建一个新的带加密的数据包读写器
func NewPacketReadWriterWithCryptor(minEncryptedLen, maxEncryptedLen uint32, timeout time.Duration) *PacketReadWriterWithCryptor {
	return &PacketReadWriterWithCryptor{
		PacketReadWriter: NewPacketReadWriter(minEncryptedLen, maxEncryptedLen, timeout),
	}
}

// ReadAndDecryptPacket 读取并解密数据包
func (prw *PacketReadWriterWithCryptor) ReadAndDecryptPacket(conn net.Conn, decryptor crypto.Decryptor) ([]byte, error) {
	// 读取数据包
	p, err := prw.ReadPacket(conn)
	if err != nil {
		return nil, err
	}

	// 解密数据包
	decrypted, err := decryptor.Decrypt(p)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "decrypt packet")
	}

	return decrypted, nil
}

// EncryptAndWritePacket 加密并写入数据包
func (prw *PacketReadWriterWithCryptor) EncryptAndWritePacket(conn net.Conn, p []byte, encryptor crypto.Encryptor) error {
	// 加密数据包
	encrypted, err := encryptor.Encrypt(p)
	if err != nil {
		return pkgerrors.WithMessage(err, "encrypt packet")
	}

	// 写入加密后的数据包
	return prw.PacketReadWriter.WritePacket(conn, encrypted)
}
