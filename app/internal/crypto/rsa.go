package crypto

import (
	"github.com/godyy/ggs/internal/core/crypto"
	"github.com/godyy/ggs/internal/core/crypto/rsa"
)

const priKey = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCxd3WTDARg9SVs
ALqjp2Z9qN+4i0QsXnTAP2YzC7eUmBYco6Zeg4K64pKYR+a02cC5xkM8MIVvXFDJ
uQGfO31LJnenVEqkVRCCugpPVfadQLixfN1GWgAujqLDuKgZfkViEezkmK75Gy1W
vdw1bolepWoHk+7+5PgCKIr1cojk8l+l2nFtqTzjKhObBxnbzmES34Z754g+emZq
X9KmQwPd2FMfi4kfSvT4f2MfY6Fbzqi9f84eGqxAgUQIMApWN4T6EvlNHF64VM5M
tuiu+LeE+O5ElC1r+7FSLoRRigLNdU5Dojw+IwC414Kr8X84B5d2Pafe9wFFD3FA
TI0Youd3AgMBAAECggEAL6trBajAtFqlRrGbcMJSoYZvMd8W3OQycEGXbjbDhUKl
DeRXmCOzRgf+YLFPo1yqjDxZax2NejBN8yGi8ebE7R7UHTpjImlHGhZnFpB8whjU
g7iKp48dZWQjDHfZj59/e6xc+bqZpYhLUXWGZUPf2nCMXqS6GfXfOJUzXmL5qqWN
vyGtWL94cDiSTWJ1AZycbGLwIml/rZ1uOB513hZmbuzDT6OusECxzo5aW3y+7KTC
ZP7b2AyREQioHq+eplC2iqvvP3n4zMBwo+/XpNEEhu7g5HVAvP97arfGR00fkxo9
avPVS2GlYqLxNzkTIn6AxvFzLFeAxNb1fHFJ7lObgQKBgQDczy1+haJjmTvWBFIK
DCtqcgI9Tn/xTtmc5dKLkhRhTO1ykWJ2kghv5ZfcW/9HesBws3la4SdoKbCCNVAP
nzCp7Bxf1XG+a5xqDZV6GTOpqgS70yIGXVOk8bwXcSxlA7qR6Yl6PQg2eSfHHBZQ
rXMuVHlP9t0pkF4cRCv04qiC0QKBgQDNv/CmcANXZx3YDllacL11a+HQv6zHotRT
tC6bw6xFiQSr6ClyH63efWilvYeuuhZRfdIWEWBB+MaArziqw14nfz1c3UTxhYhw
v1PBABeiTWFS6S8RGE8pI2G3Z2I6X+T1ZeMBpvVvRNvTnLx+H0sDMxGeYYSfvh62
mblPkFmHxwKBgDdFsRirsOOHlv/Sowqa0z9Y/JCGFua7myN4MAT58xoMHKACHoiZ
s3z3FtV1PeiRpJxRgL4sACZF0UY2vCy853yReuTOVCObYlL1xYYDyvfcdETj6+91
6xst26xuivNaRJiDwgMURfsExt1DfZ6CXIOrZ5aJsADYf4ZJ1kr9dbsRAoGBALw2
e/7U8smOg6d0INrxzO5QPOcHoBeTZWYYqpZE3h9R4xsaqmdCgXvI/uS2xxrYEbiE
P51+Ua6n03Y+U7kqNMQuykRcCUhjHdf9vbEM05Hd9UyyESMzOJ7qReZPRXUe6cRu
asXFJDmgJPOkKm25VJZdrh1TGc5DTbc+Ul1tL+lbAoGAHu0/hwbU3l24WsOZVfYn
DgiMHyUuBuv9W87OPHweiYmR4FgnGk3ULu+oufTv00cH91+OMpw/mF63tC024dDE
CjQ06a1/KkrzbuddbCF1p8UK5TXMmWWDZzNwrhiSe7c22VkBrKZ1WGG0VE1BKhjA
BqIMHd6m95nxo1Bg6OSa1hY=
-----END PRIVATE KEY-----
`

const pubKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAsXd1kwwEYPUlbAC6o6dm
fajfuItELF50wD9mMwu3lJgWHKOmXoOCuuKSmEfmtNnAucZDPDCFb1xQybkBnzt9
SyZ3p1RKpFUQgroKT1X2nUC4sXzdRloALo6iw7ioGX5FYhHs5Jiu+RstVr3cNW6J
XqVqB5Pu/uT4AiiK9XKI5PJfpdpxbak84yoTmwcZ285hEt+Ge+eIPnpmal/SpkMD
3dhTH4uJH0r0+H9jH2OhW86ovX/OHhqsQIFECDAKVjeE+hL5TRxeuFTOTLborvi3
hPjuRJQta/uxUi6EUYoCzXVOQ6I8PiMAuNeCq/F/OAeXdj2n3vcBRQ9xQEyNGKLn
dwIDAQAB
-----END PUBLIC KEY-----`

// CreateRSAEncryptor 创建RSA加密器.
func CreateRSAEncryptor() (crypto.Entryptor, error) {
	pubKey, err := rsa.ParsePublicKeyFromPEM([]byte(pubKey))
	if err != nil {
		return nil, err
	}
	return rsa.NewRSAEncryptor(pubKey), nil
}

// CreateRSADecryptor 创建RSA解密器.
func CreateRSADecryptor() (crypto.Decryptor, error) {
	priKey, err := rsa.ParsePrivateKeyFromPEM([]byte(priKey))
	if err != nil {
		return nil, err
	}
	return rsa.NewRSADecryptor(priKey), nil
}
