package bytes

import (
	"encoding/binary"
	"errors"
	"math"
)

// ErrVarintOverflow  varint值溢出
var ErrVarintOverflow = errors.New("bytes: varint overflow")

const (
	MaxStringLength = math.MaxInt32 // maximum string length
)

// MaxVarintLenN is the maximum length of a varint-encoded N-bit integer.
const (
	MaxVarintLen16 = binary.MaxVarintLen16
	MaxVarintLen32 = binary.MaxVarintLen32
	MaxVarintLen64 = binary.MaxVarintLen64
)

// MaxStringLenLen maximum length of string
const MaxStringLenLen = MaxVarintLen32

// byteOrder 字节序
type byteOrder = binary.ByteOrder

var (
	nativeEndian = &binary.NativeEndian // 本地字节序
	littleEndian = &binary.LittleEndian // 小端字节序
	bigEndian    = &binary.BigEndian    // 大端字节序
)
