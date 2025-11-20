package crypto

import (
	"github.com/godyy/ggs/internal/base/crypto/aes"
)

// CreateAESCrypto 创建AES对称加密工具
func CreateAESCrypto(secretKey []byte) (aes.Cryptor, error) {
	return aes.NewAESGCMCryptor(secretKey, secretKey[:12])
}
