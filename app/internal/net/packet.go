package net

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/godyy/ggs/internal/consts"
	"github.com/godyy/ggs/internal/core/crypto"
	codecc2s "github.com/godyy/ggs/internal/proto/codec/c2s"
	pkgerrors "github.com/pkg/errors"
)

// ReadPacket 读取数据包.
func ReadPacket(conn net.Conn) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(consts.ReadWriteTimeout))

	var l [4]byte
	if _, err := io.ReadFull(conn, l[:]); err != nil {
		return nil, pkgerrors.WithMessage(err, "read length")
	}

	n := binary.BigEndian.Uint32(l[:])
	if n < codecc2s.HeadLen {
		return nil, fmt.Errorf("invalid packet len %d", n)
	}

	p := make([]byte, n)
	if _, err := io.ReadFull(conn, p); err != nil {
		return nil, pkgerrors.WithMessage(err, "read payload")
	}

	return p, nil
}

// writeExact 确保写入指定数量的字节
func writeExact(conn net.Conn, data []byte) error {
	total := len(data)
	written := 0

	for written < total {
		n, err := conn.Write(data[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}

// WritePacket 写入数据包.
func WritePacket(conn net.Conn, p []byte) error {
	conn.SetWriteDeadline(time.Now().Add(consts.ReadWriteTimeout))

	var l [4]byte
	binary.BigEndian.PutUint32(l[:], uint32(len(p)))

	// 确保完整写入4字节长度字段
	if err := writeExact(conn, l[:]); err != nil {
		return pkgerrors.WithMessage(err, "write length")
	}

	// 确保完整写入数据包
	if err := writeExact(conn, p); err != nil {
		return pkgerrors.WithMessage(err, "write packet")
	}

	return nil
}

// ReadAndDecryptPacket 读取并解密数据包.
func ReadAndDecryptPacket(conn net.Conn, decryptor crypto.Decryptor) ([]byte, error) {
	// 读取数据包
	p, err := ReadPacket(conn)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "read packet")
	}

	// 解密数据包
	decrypted, err := decryptor.Decrypt(p)
	if err != nil {
		return nil, pkgerrors.WithMessage(err, "decrypt packet")
	}

	return decrypted, nil
}

// EncryptAndWritePacket 加密并写入数据包.
func EncryptAndWritePacket(conn net.Conn, p []byte, encryptor crypto.Entryptor) error {
	// 加密数据包
	encrypted, err := encryptor.Encrypt(p)
	if err != nil {
		return pkgerrors.WithMessage(err, "encrypt packet")
	}

	// 写入加密后的数据包
	if err := WritePacket(conn, encrypted); err != nil {
		return pkgerrors.WithMessage(err, "write packet")
	}

	return nil
}
