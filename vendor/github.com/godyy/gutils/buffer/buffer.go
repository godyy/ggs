package buffer

import "errors"

// ErrBufferFull 缓冲区已满
var ErrBufferFull = errors.New("buffer is full")

// ErrExceedBufferLimit 超过缓冲区长度限制
var ErrExceedBufferLimit = errors.New("exceed buffer limit")

// ErrStringLenExceedLimit 字符串长度超过限制
var ErrStringLenExceedLimit = errors.New("string length exceed limit")
