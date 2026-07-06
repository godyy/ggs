package bytes

import (
	"github.com/godyy/gutils/buffer"
	"io"
	"math"
)

func (b *FixedBuffer) readInt16(bo byteOrder) (int16, error) {
	n, err := b.readUint16(bo)
	return int16(n), err
}

func (b *FixedBuffer) writeInt16(i int16, bo byteOrder) error {
	return b.writeUint16(uint16(i), bo)
}

func (b *FixedBuffer) readUint16(bo byteOrder) (i uint16, err error) {
	l := b.Readable()
	if l == 0 {
		return 0, io.EOF
	}
	if l < 2 {
		return 0, io.ErrUnexpectedEOF
	}

	i = bo.Uint16(b.buf[b.r : b.r+2])
	b.r += 2
	return
}

func (b *FixedBuffer) writeUint16(i uint16, bo byteOrder) error {
	l := b.Writable()
	if l == 0 {
		return buffer.ErrBufferFull
	}
	if l < 2 {
		return buffer.ErrExceedBufferLimit
	}

	b.slideReadable()

	bo.PutUint16(b.buf[b.w:b.w+2], i)
	b.w += 2
	return nil
}

func (b *FixedBuffer) readInt32(bo byteOrder) (int32, error) {
	n, err := b.readUint32(bo)
	return int32(n), err
}

func (b *FixedBuffer) writeInt32(i int32, bo byteOrder) error {
	return b.writeUint32(uint32(i), bo)
}

func (b *FixedBuffer) readUint32(bo byteOrder) (i uint32, err error) {
	l := b.Readable()
	if l == 0 {
		return 0, io.EOF
	}
	if l < 4 {
		return 0, io.ErrUnexpectedEOF
	}

	i = bo.Uint32(b.buf[b.r : b.r+4])
	b.r += 4
	return
}

func (b *FixedBuffer) writeUint32(i uint32, bo byteOrder) error {
	l := b.Writable()
	if l == 0 {
		return buffer.ErrBufferFull
	}
	if l < 4 {
		return buffer.ErrExceedBufferLimit
	}

	b.slideReadable()

	bo.PutUint32(b.buf[b.w:b.w+4], i)
	b.w += 4
	return nil
}

func (b *FixedBuffer) readInt64(bo byteOrder) (int64, error) {
	n, err := b.readUint64(bo)
	return int64(n), err
}

func (b *FixedBuffer) writeInt64(i int64, bo byteOrder) error {
	return b.writeUint64(uint64(i), bo)
}

func (b *FixedBuffer) readUint64(bo byteOrder) (i uint64, err error) {
	l := b.Readable()
	if l == 0 {
		return 0, io.EOF
	}
	if l < 8 {
		return 0, io.ErrUnexpectedEOF
	}

	i = bo.Uint64(b.buf[b.r : b.r+8])
	b.r += 8
	return
}

func (b *FixedBuffer) writeUint64(i uint64, bo byteOrder) error {
	l := b.Writable()
	if l == 0 {
		return buffer.ErrBufferFull
	}
	if l < 8 {
		return buffer.ErrExceedBufferLimit
	}

	b.slideReadable()

	bo.PutUint64(b.buf[b.w:b.w+8], i)
	b.w += 8
	return nil
}

func (b *FixedBuffer) writeFloat32(f float32, bo byteOrder) error {
	var buf [4]byte
	bo.PutUint32(buf[:], math.Float32bits(f))
	_, err := b.Write(buf[:])
	return err
}

func (b *FixedBuffer) readFloat32(bo byteOrder) (float32, error) {
	var buf [4]byte
	if _, err := b.Read(buf[:]); err != nil {
		return 0, err
	}
	return math.Float32frombits(bo.Uint32(buf[:])), nil
}

func (b *FixedBuffer) writeFloat64(f float64, bo byteOrder) error {
	var buf [8]byte
	bo.PutUint64(buf[:], math.Float64bits(f))
	_, err := b.Write(buf[:])
	return err
}

func (b *FixedBuffer) readFloat64(bo byteOrder) (float64, error) {
	var buf [8]byte
	if _, err := b.Read(buf[:]); err != nil {
		return 0, err
	}
	return math.Float64frombits(bo.Uint64(buf[:])), nil
}

func (b *FixedBuffer) ReadBigInt16() (int16, error) {
	return b.readInt16(bigEndian)
}

func (b *FixedBuffer) WriteBigInt16(i int16) error {
	return b.writeInt16(i, bigEndian)
}

func (b *FixedBuffer) ReadBigUint16() (i uint16, err error) {
	return b.readUint16(bigEndian)
}

func (b *FixedBuffer) WriteBigUint16(i uint16) error {
	return b.writeUint16(i, bigEndian)
}

func (b *FixedBuffer) ReadBigInt32() (int32, error) {
	return b.readInt32(bigEndian)
}

func (b *FixedBuffer) WriteBigInt32(i int32) error {
	return b.writeInt32(i, bigEndian)
}

func (b *FixedBuffer) ReadBigUint32() (i uint32, err error) {
	return b.readUint32(bigEndian)
}

func (b *FixedBuffer) WriteBigUint32(i uint32) error {
	return b.writeUint32(i, bigEndian)
}

func (b *FixedBuffer) ReadBigInt64() (int64, error) {
	return b.readInt64(bigEndian)
}

func (b *FixedBuffer) WriteBigInt64(i int64) error {
	return b.writeInt64(i, bigEndian)
}

func (b *FixedBuffer) ReadBigUint64() (i uint64, err error) {
	return b.readUint64(bigEndian)
}

func (b *FixedBuffer) WriteBigUint64(i uint64) error {
	return b.writeUint64(i, bigEndian)
}

func (b *FixedBuffer) WriteBigFloat32(f float32) error {
	return b.writeFloat32(f, bigEndian)
}

func (b *FixedBuffer) ReadBigFloat32() (f float32, err error) {
	return b.readFloat32(bigEndian)
}

func (b *FixedBuffer) WriteBigFloat64(f float64) error {
	return b.writeFloat64(f, bigEndian)
}

func (b *FixedBuffer) ReadBigFloat64() (f float64, err error) {
	return b.readFloat64(bigEndian)
}

func (b *FixedBuffer) ReadLitInt16() (int16, error) {
	return b.readInt16(littleEndian)
}

func (b *FixedBuffer) WriteLitInt16(i int16) error {
	return b.writeInt16(i, littleEndian)
}

func (b *FixedBuffer) ReadLitUint16() (i uint16, err error) {
	return b.readUint16(littleEndian)
}

func (b *FixedBuffer) WriteLitUint16(i uint16) error {
	return b.writeUint16(i, littleEndian)
}

func (b *FixedBuffer) ReadLitInt32() (int32, error) {
	return b.readInt32(littleEndian)
}

func (b *FixedBuffer) WriteLitInt32(i int32) error {
	return b.writeInt32(i, littleEndian)
}

func (b *FixedBuffer) ReadLitUint32() (i uint32, err error) {
	return b.readUint32(littleEndian)
}

func (b *FixedBuffer) WriteLitUint32(i uint32) error {
	return b.writeUint32(i, littleEndian)
}

func (b *FixedBuffer) ReadLitInt64() (int64, error) {
	return b.readInt64(littleEndian)
}

func (b *FixedBuffer) WriteLitInt64(i int64) error {
	return b.writeInt64(i, littleEndian)
}

func (b *FixedBuffer) ReadLitUint64() (i uint64, err error) {
	return b.readUint64(littleEndian)
}

func (b *FixedBuffer) WriteLitUint64(i uint64) error {
	return b.writeUint64(i, littleEndian)
}

func (b *FixedBuffer) WriteLitFloat32(f float32) error {
	return b.writeFloat32(f, littleEndian)
}

func (b *FixedBuffer) ReadLitFloat32() (f float32, err error) {
	return b.readFloat32(littleEndian)
}

func (b *FixedBuffer) WriteLitFloat64(f float64) error {
	return b.writeFloat64(f, littleEndian)
}

func (b *FixedBuffer) ReadLitFloat64() (f float64, err error) {
	return b.readFloat64(littleEndian)
}
