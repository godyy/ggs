package bytes

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/godyy/gutils/buffer"
	pkg_errors "github.com/pkg/errors"
)

// ErrBufferTooLarge is passed to panic if memory cannot be allocated to store db in a buffer.
var ErrBufferTooLarge = errors.New("bytes: buffer too large")

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64

const maxInt = int(^uint(0) >> 1)

// Buffer 变长byte缓冲区
type Buffer struct {
	buf []byte
	off int
}

// NewBufferWithCap 已指定的容量创建Buffer
func NewBufferWithCap(cap int) *Buffer {
	if cap < smallBufferSize {
		cap = smallBufferSize
	}
	return &Buffer{
		buf: make([]byte, 0, cap),
		off: 0,
	}
}

// NewBuffer 指定buf创建Buffer
func NewBuffer(buf []byte) *Buffer {
	if buf == nil {
		buf = make([]byte, 0, smallBufferSize)
	}

	b := &Buffer{
		buf: buf,
		off: 0,
	}
	return b
}

// Size 获取buf的大小
func (b *Buffer) Size() int {
	return len(b.buf)
}

// Cap 获取buf的容量
func (b *Buffer) Cap() int {
	return cap(b.buf)
}

// Reset 重置buf
func (b *Buffer) Reset() {
	if b.buf != nil {
		b.buf = b.buf[:0]
	}
	b.off = 0
}

// SetBuf 设置buf并返回之前的buf
func (b *Buffer) SetBuf(buf []byte) []byte {
	old := b.buf
	b.buf = buf
	b.off = 0
	return old
}

// Readable 获取可读取字节数
func (b *Buffer) Readable() int {
	return len(b.buf) - b.off
}

// Writable 获取根据当前容量还可写入的字节数
func (b *Buffer) Writable() int {
	return cap(b.buf) - len(b.buf)
}

// Data 获取buf中的完整数据
func (b *Buffer) Data() []byte {
	if b.buf == nil {
		return nil
	}
	return b.buf[:]
}

// UnreadData 获取buf中的未读数据
func (b *Buffer) UnreadData() []byte {
	if b.buf == nil {
		return nil
	}
	return b.buf[b.off:]
}

// tryGrowByReslice 尝试扩张buf
func (b *Buffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// growBuffSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
func growBuffSlice(b []byte, n int) []byte {
	defer func() {
		if recover() != nil {
			panic(ErrBufferTooLarge)
		}
	}()
	// TODO(http://golang.org/issue/51462): We should rely on the append-make
	// pattern so that the compiler can call runtime.growslice. For example:
	//	return append(b, make([]byte, n)...)
	// This avoids unnecessary zero-ing of the first len(b) bytes of the
	// allocated slice, but this pattern causes b to escape onto the heap.
	//
	// Instead use the append-make pattern with a nil slice to ensure that
	// we allocate buffers rounded up to the closest size class.
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		// The growth rate has historically always been 2x. In the future,
		// we could rely purely on append to determine the growth rate.
		c = 2 * cap(b)
	}
	b2 := append([]byte(nil), make([]byte, c)...)
	copy(b2, b)
	return b2[:len(b)]
}

// grow 扩张buf使其能容纳n个字节
func (b *Buffer) grow(n int) int {
	m := b.Readable()
	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Reset()
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]byte, n, smallBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrBufferTooLarge)
	} else {
		// Add b.off to account for b.buf[:b.off] being sliced off the front.
		b.buf = growBuffSlice(b.buf[b.off:], b.off+n)
	}
	// Restore b.off and len(b.buf).
	b.off = 0
	b.buf = b.buf[:m+n]

	return m
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with ErrTooLarge.
// If size is true, it grow the buffer's size.
func (b *Buffer) Grow(n int, size ...bool) {
	if n < 0 {
		panic("bytes.Buffer.Grow: n < 0")
	}
	m := b.grow(n)
	if len(size) <= 0 || !size[0] {
		b.buf = b.buf[:m]
	}
}

func (b *Buffer) ReadByte() (c byte, err error) {
	if b.Readable() == 0 {
		return 0, io.EOF
	}

	c = b.buf[b.off]
	b.off++
	return
}

func (b *Buffer) WriteByte(c byte) error {
	m, ok := b.tryGrowByReslice(1)
	if !ok {
		m = b.grow(1)
	}
	b.buf[m] = c
	return nil
}

func (b *Buffer) ReadInt8() (int8, error) {
	n, err := b.ReadUint8()
	return int8(n), err
}

func (b *Buffer) WriteInt8(i int8) error {
	return b.WriteUint8(uint8(i))
}

func (b *Buffer) ReadUint8() (i uint8, err error) {
	return b.ReadByte()
}

func (b *Buffer) WriteUint8(i uint8) error {
	return b.WriteByte(i)
}

func (b *Buffer) ReadInt16() (int16, error) {
	return b.readInt16(nativeEndian)
}

func (b *Buffer) WriteInt16(i int16) error {
	return b.writeInt16(i, nativeEndian)
}

func (b *Buffer) ReadUint16() (uint16, error) {
	return b.readUint16(nativeEndian)
}

func (b *Buffer) WriteUint16(i uint16) error {
	return b.writeUint16(i, nativeEndian)
}

func (b *Buffer) ReadInt32() (int32, error) {
	return b.readInt32(nativeEndian)
}

func (b *Buffer) WriteInt32(i int32) error {
	return b.writeInt32(i, nativeEndian)
}

func (b *Buffer) ReadUint32() (uint32, error) {
	return b.readUint32(nativeEndian)
}

func (b *Buffer) WriteUint32(i uint32) error {
	return b.writeUint32(i, nativeEndian)
}

func (b *Buffer) ReadInt64() (int64, error) {
	return b.readInt64(nativeEndian)
}

func (b *Buffer) WriteInt64(i int64) error {
	return b.writeInt64(i, nativeEndian)
}

func (b *Buffer) ReadUint64() (uint64, error) {
	return b.readUint64(nativeEndian)
}

func (b *Buffer) WriteUint64(i uint64) error {
	return b.writeUint64(i, nativeEndian)
}

func (b *Buffer) ReadFloat32() (float32, error) {
	return b.readFloat32(nativeEndian)
}

func (b *Buffer) WriteFloat32(f float32) error {
	return b.writeFloat32(f, nativeEndian)
}

func (b *Buffer) ReadFloat64() (float64, error) {
	return b.readFloat64(nativeEndian)
}

func (b *Buffer) WriteFloat64(f float64) error {
	return b.writeFloat64(f, nativeEndian)
}

func (b *Buffer) ReadBool() (bool, error) {
	c, err := b.ReadByte()
	if err != nil {
		return false, err
	}
	return c == 1, nil
}

func (b *Buffer) WriteBool(v bool) error {
	if v {
		return b.WriteByte(1)
	} else {
		return b.WriteByte(0)
	}
}

func (b *Buffer) ReadVarint16() (int16, error) {
	i, n := binary.Varint(b.buf[b.off:])
	if n == 0 {
		return 0, io.EOF
	}
	if n < 0 || n > MaxVarintLen16 {
		return 0, ErrVarintOverflow
	}
	b.off += n
	return int16(i), nil
}

func (b *Buffer) WriteVarint16(i int16) (int, error) {
	var buf [MaxVarintLen16]byte
	n := binary.PutVarint(buf[:], int64(i))

	m, ok := b.tryGrowByReslice(n)
	if !ok {
		m = b.grow(n)
	}

	copy(b.buf[m:m+n], buf[:n])
	return n, nil
}

func (b *Buffer) ReadUvarint16() (uint16, error) {
	i, n := binary.Uvarint(b.buf[b.off:])
	if n == 0 {
		return 0, io.EOF
	}
	if n < 0 || n > MaxVarintLen16 {
		return 0, ErrVarintOverflow
	}
	b.off += n
	return uint16(i), nil
}

func (b *Buffer) WriteUvarint16(i uint16) (int, error) {
	var buf [MaxVarintLen16]byte
	n := binary.PutUvarint(buf[:], uint64(i))

	m, ok := b.tryGrowByReslice(n)
	if !ok {
		m = b.grow(n)
	}

	copy(b.buf[m:m+n], buf[:n])
	return n, nil
}

func (b *Buffer) ReadVarint32() (int32, error) {
	i, n := binary.Varint(b.buf[b.off:])
	if n == 0 {
		return 0, io.EOF
	}
	if n < 0 || n > MaxVarintLen32 {
		return 0, ErrVarintOverflow
	}
	b.off += n
	return int32(i), nil
}

func (b *Buffer) WriteVarint32(i int32) (int, error) {
	var buf [MaxVarintLen32]byte
	n := binary.PutVarint(buf[:], int64(i))

	m, ok := b.tryGrowByReslice(n)
	if !ok {
		m = b.grow(n)
	}

	copy(b.buf[m:m+n], buf[:n])
	return n, nil
}

func (b *Buffer) ReadUvarint32() (uint32, error) {
	i, n := binary.Uvarint(b.buf[b.off:])
	if n == 0 {
		return 0, io.EOF
	}
	if n < 0 || n > MaxVarintLen32 {
		return 0, ErrVarintOverflow
	}
	b.off += n
	return uint32(i), nil
}

func (b *Buffer) WriteUvarint32(i uint32) (int, error) {
	var buf [MaxVarintLen32]byte
	n := binary.PutUvarint(buf[:], uint64(i))

	m, ok := b.tryGrowByReslice(n)
	if !ok {
		m = b.grow(n)
	}

	copy(b.buf[m:m+n], buf[:n])
	return n, nil
}

func (b *Buffer) ReadVarint64() (int64, error) {
	i, n := binary.Varint(b.buf[b.off:])
	if n == 0 {
		return 0, io.EOF
	}
	if n < 0 {
		return 0, ErrVarintOverflow
	}
	b.off += n
	return i, nil
}

func (b *Buffer) WriteVarint64(i int64) (int, error) {
	var buf [MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], i)

	m, ok := b.tryGrowByReslice(n)
	if !ok {
		m = b.grow(n)
	}

	copy(b.buf[m:m+n], buf[:n])
	return n, nil
}

func (b *Buffer) ReadUvarint64() (uint64, error) {
	i, n := binary.Uvarint(b.buf[b.off:])
	if n == 0 {
		return 0, io.EOF
	}
	if n < 0 {
		return 0, ErrVarintOverflow
	}
	b.off += n
	return i, nil
}

func (b *Buffer) WriteUvarint64(i uint64) (int, error) {
	var buf [MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], i)

	m, ok := b.tryGrowByReslice(n)
	if !ok {
		m = b.grow(n)
	}

	copy(b.buf[m:m+n], buf[:n])
	return n, nil
}

func (b *Buffer) Read(p []byte) (int, error) {
	if len(p) <= 0 {
		return 0, io.ErrShortBuffer
	}

	if b.Readable() == 0 {
		return 0, io.EOF
	}

	n := copy(p, b.buf[b.off:])
	b.off += n
	return n, nil
}

func (b *Buffer) Write(p []byte) (n int, err error) {
	m, ok := b.tryGrowByReslice(len(p))
	if !ok {
		m = b.grow(len(p))
	}
	return copy(b.buf[m:], p), nil
}

func (b *Buffer) ReadString() (string, error) {
	i, n := binary.Varint(b.buf[b.off:])
	if n == 0 {
		return "", io.EOF
	}
	if n < 0 || n > MaxStringLenLen {
		return "", pkg_errors.WithMessage(ErrVarintOverflow, "read length")
	}

	l := int(i)
	if l < 0 {
		return "", buffer.ErrStringLenExceedLimit
	}

	if l+n > b.Readable() {
		return "", io.ErrUnexpectedEOF
	}

	b.off += n
	if l == 0 {
		return "", nil
	}
	s := string(b.buf[b.off : b.off+l])
	b.off += l
	return s, nil
}

func (b *Buffer) WriteString(s string) error {
	l := len(s)
	if l > MaxStringLength {
		return buffer.ErrStringLenExceedLimit
	}

	var buf [MaxStringLenLen]byte
	ll := binary.PutVarint(buf[:], int64(l))

	m, ok := b.tryGrowByReslice(ll + l)
	if !ok {
		m = b.grow(ll + l)
	}

	copy(b.buf[m:m+ll], buf[:ll])
	m = m + ll
	copy(b.buf[m:m+l], s)
	return nil
}

func (b *Buffer) ReadFrom(r io.Reader) (int64, error) {
	l := len(b.buf)
	c := cap(b.buf)
	if l >= c {
		return 0, buffer.ErrBufferFull
	}

	return b.ReadFromN(r, c-l)
}

func (b *Buffer) ReadFromN(r io.Reader, n int) (int64, error) {
	if n < 0 {
		panic("bytes.Buffer.ReadFromN: n < 0")
	}

	m := b.grow(n)
	b.buf = b.buf[:m]

	nn, err := r.Read(b.buf[m : m+n])
	if nn < 0 {
		panic("bytes.Buffer.ReadFromN: reader returned negative count from Read")
	}

	b.buf = b.buf[:m+nn]
	return int64(nn), err
}

func (b *Buffer) WriteTo(w io.Writer) (int64, error) {
	l := b.Readable()
	if l == 0 {
		return 0, nil
	}

	n, err := w.Write(b.buf[b.off:])
	if n > l {
		panic("bytes.Buffer: invalid Write count")
	}

	b.off += n
	return int64(n), err
}

// Peek 自buf中提取n个字节，同时并不会更新已读偏移
func (b *Buffer) Peek(n int) ([]byte, error) {
	if n > len(b.buf) {
		return nil, buffer.ErrExceedBufferLimit
	}

	if n > b.Readable() {
		return b.buf[b.off:], io.ErrUnexpectedEOF
	}

	return b.buf[b.off : b.off+n], nil
}

// Skip 自buf中跳过n个字节
func (b *Buffer) Skip(n int) (skipped int, err error) {
	if n > len(b.buf) {
		return 0, buffer.ErrExceedBufferLimit
	}

	if l := b.Readable(); n > l {
		skipped = l
		err = io.ErrUnexpectedEOF
	} else {
		skipped = n
	}

	b.off += skipped
	return
}
