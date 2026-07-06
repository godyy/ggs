package gnet

import (
	"io"
	"net"
	"time"
)

// ConnReader 连接Reader
type ConnReader interface {
	SetReadDeadline(t time.Time) error
	io.Reader
}

// ConnWriter 连接Writer
type ConnWriter interface {
	SetWriteDeadline(t time.Time) error
	io.Writer
}

// ConnReaderFrom 连接ReaderFrom
type ConnReaderFrom interface {
	SetReadDeadline(t time.Time) error
	ReadFrom([]byte) (int, net.Addr, error)
}

// ConnWriterTo 连接WriterTo
type ConnWriterTo interface {
	SetWriteDeadline(t time.Time) error
	WriteTo([]byte, net.Addr) (int, error)
}

// ReadFull 自 conn 中读取 p 长度的数据.
func ReadFull(conn net.Conn, p []byte) error {
	for len(p) > 0 {
		n, err := conn.Read(p)
		if err != nil {
			return err
		}
		p = p[n:]
	}
	return nil
}

// WriteFull 将 p 长度的数据写入 conn.
func WriteFull(conn net.Conn, p []byte) error {
	for len(p) > 0 {
		n, err := conn.Write(p)
		if err != nil {
			return err
		}
		p = p[n:]
	}
	return nil
}
