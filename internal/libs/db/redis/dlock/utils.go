package dlock

import (
	"crypto/rand"
	"fmt"
	"io"
	rand2 "math/rand"
)

// GenerateUUID 生成一个简单的UUID.
func GenerateUUID() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return fmt.Sprintf("%d", rand2.Int63())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant is 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
