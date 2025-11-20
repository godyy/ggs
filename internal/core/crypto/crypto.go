package crypto

// Entryptor 加密器
type Entryptor interface {
	Encrypt(data []byte) ([]byte, error)
}

// Decryptor 解密器
type Decryptor interface {
	Decrypt(data []byte) ([]byte, error)
}

// Signer 签名器
type Signer interface {
	Sign(data []byte) ([]byte, error)
}

// Verifier 验签器
type Verifier interface {
	Verify(data []byte, signature []byte) error
}
