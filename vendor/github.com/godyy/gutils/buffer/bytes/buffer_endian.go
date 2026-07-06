package bytes

import (
	"io"
	"math"
)

func (b *Buffer) readInt16(bo byteOrder) (int16, error) {
	n, err := b.readUint16(bo)
	return int16(n), err
}

func (b *Buffer) writeInt16(i int16, bo byteOrder) error {
	return b.writeUint16(uint16(i), bo)
}

func (b *Buffer) readUint16(bo byteOrder) (i uint16, err error) {
	l := b.Readable()
	if l == 0 {
		return 0, io.EOF
	}
	if l < 2 {
		return 0, io.ErrUnexpectedEOF
	}

	i = bo.Uint16(b.buf[b.off : b.off+2])
	b.off += 2
	return
}

func (b *Buffer) writeUint16(i uint16, bo byteOrder) error {
	m, ok := b.tryGrowByReslice(2)
	if !ok {
		m = b.grow(2)
	}
	bo.PutUint16(b.buf[m:m+2], i)
	return nil
}

func (b *Buffer) readInt32(bo byteOrder) (int32, error) {
	n, err := b.readUint32(bo)
	return int32(n), err
}

func (b *Buffer) writeInt32(i int32, bo byteOrder) error {
	return b.writeUint32(uint32(i), bo)
}

func (b *Buffer) readUint32(bo byteOrder) (i uint32, err error) {
	l := b.Readable()
	if l == 0 {
		return 0, io.EOF
	}
	if l < 4 {
		return 0, io.ErrUnexpectedEOF
	}

	i = bo.Uint32(b.buf[b.off : b.off+4])
	b.off += 4
	return
}

func (b *Buffer) writeUint32(i uint32, bo byteOrder) error {
	m, ok := b.tryGrowByReslice(4)
	if !ok {
		m = b.grow(4)
	}
	bo.PutUint32(b.buf[m:m+4], i)
	return nil
}

func (b *Buffer) readInt64(bo byteOrder) (int64, error) {
	n, err := b.readUint64(bo)
	return int64(n), err
}

func (b *Buffer) writeInt64(i int64, bo byteOrder) error {
	return b.writeUint64(uint64(i), bo)
}

func (b *Buffer) readUint64(bo byteOrder) (i uint64, err error) {
	l := b.Readable()
	if l == 0 {
		return 0, io.EOF
	}
	if l < 8 {
		return 0, io.ErrUnexpectedEOF
	}

	i = bo.Uint64(b.buf[b.off : b.off+8])
	b.off += 8
	return
}

func (b *Buffer) writeUint64(i uint64, bo byteOrder) error {
	m, ok := b.tryGrowByReslice(8)
	if !ok {
		m = b.grow(8)
	}
	bo.PutUint64(b.buf[m:m+8], i)
	return nil
}

func (b *Buffer) readFloat32(bo byteOrder) (float32, error) {
	var buf [4]byte
	if _, err := b.Read(buf[:]); err != nil {
		return 0, err
	}
	return math.Float32frombits(bo.Uint32(buf[:])), nil
}

func (b *Buffer) writeFloat32(f float32, bo byteOrder) error {
	var buf [4]byte
	bo.PutUint32(buf[:], math.Float32bits(f))
	_, err := b.Write(buf[:])
	return err
}

func (b *Buffer) readFloat64(bo byteOrder) (float64, error) {
	var buf [8]byte
	if _, err := b.Read(buf[:]); err != nil {
		return 0, err
	}
	return math.Float64frombits(bo.Uint64(buf[:])), nil
}

func (b *Buffer) writeFloat64(f float64, bo byteOrder) error {
	var buf [8]byte
	bo.PutUint64(buf[:], math.Float64bits(f))
	_, err := b.Write(buf[:])
	return err
}

func (b *Buffer) ReadLitInt16() (int16, error) {
	return b.readInt16(littleEndian)
}

func (b *Buffer) WriteLitInt16(i int16) error {
	return b.writeInt16(i, littleEndian)
}

func (b *Buffer) ReadLitUint16() (i uint16, err error) {
	return b.readUint16(littleEndian)
}

func (b *Buffer) WriteLitUint16(i uint16) error {
	return b.writeUint16(i, littleEndian)
}

func (b *Buffer) ReadLitInt32() (int32, error) {
	return b.readInt32(littleEndian)
}

func (b *Buffer) WriteLitInt32(i int32) error {
	return b.writeInt32(i, littleEndian)
}

func (b *Buffer) ReadLitUint32() (i uint32, err error) {
	return b.readUint32(littleEndian)
}

func (b *Buffer) WriteLitUint32(i uint32) error {
	return b.writeUint32(i, littleEndian)
}

func (b *Buffer) ReadLitInt64() (int64, error) {
	return b.readInt64(littleEndian)
}

func (b *Buffer) WriteLitInt64(i int64) error {
	return b.writeInt64(i, littleEndian)
}

func (b *Buffer) ReadLitUint64() (i uint64, err error) {
	return b.readUint64(littleEndian)
}

func (b *Buffer) WriteLitUint64(i uint64) error {
	return b.writeUint64(i, littleEndian)
}

func (b *Buffer) WriteLitFloat32(f float32) error {
	return b.writeFloat32(f, littleEndian)
}

func (b *Buffer) ReadLitFloat32() (f float32, err error) {
	return b.readFloat32(littleEndian)
}

func (b *Buffer) WriteLitFloat64(f float64) error {
	return b.writeFloat64(f, littleEndian)
}

func (b *Buffer) ReadLitFloat64() (f float64, err error) {
	return b.readFloat64(littleEndian)
}

func (b *Buffer) ReadBigInt16() (int16, error) {
	return b.readInt16(bigEndian)
}

func (b *Buffer) WriteBigInt16(i int16) error {
	return b.writeInt16(i, bigEndian)
}

func (b *Buffer) ReadBigUint16() (i uint16, err error) {
	return b.readUint16(bigEndian)
}

func (b *Buffer) WriteBigUint16(i uint16) error {
	return b.writeUint16(i, bigEndian)
}

func (b *Buffer) ReadBigInt32() (int32, error) {
	return b.readInt32(bigEndian)
}

func (b *Buffer) WriteBigInt32(i int32) error {
	return b.writeInt32(i, bigEndian)
}

func (b *Buffer) ReadBigUint32() (i uint32, err error) {
	return b.readUint32(bigEndian)
}

func (b *Buffer) WriteBigUint32(i uint32) error {
	return b.writeUint32(i, bigEndian)
}

func (b *Buffer) ReadBigInt64() (int64, error) {
	return b.readInt64(bigEndian)
}

func (b *Buffer) WriteBigInt64(i int64) error {
	return b.writeInt64(i, bigEndian)
}

func (b *Buffer) ReadBigUint64() (i uint64, err error) {
	return b.readUint64(bigEndian)
}

func (b *Buffer) WriteBigUint64(i uint64) error {
	return b.writeUint64(i, bigEndian)
}

func (b *Buffer) WriteBigFloat32(f float32) error {
	return b.writeFloat32(f, bigEndian)
}

func (b *Buffer) ReadBigFloat32() (f float32, err error) {
	return b.readFloat32(bigEndian)
}

func (b *Buffer) WriteBigFloat64(f float64) error {
	return b.writeFloat64(f, bigEndian)
}

func (b *Buffer) ReadBigFloat64() (f float64, err error) {
	return b.readFloat64(bigEndian)
}
