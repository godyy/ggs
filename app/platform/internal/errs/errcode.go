package errs

import (
	"fmt"

	"github.com/godyy/ggs/internal/errs"
)

// ErrCode 错误码
type ErrCode errs.ErrCode

func (ec ErrCode) Error() string {
	return fmt.Sprintf("%d: %s", ec, ec.Msg())
}

// Code 返回错误码
func (ec ErrCode) Code() errs.ErrCode {
	return errs.ErrCode(ec)
}

// Msg 返回错误信息
func (ec ErrCode) Msg() string {
	return errCodeStrings[ec]
}

// 错误码
const (
	ErrCodeOK            = ErrCode(0) // OK
	ErrCodeInternalError = ErrCode(1) // 服务器内部错误
)

// ErrCodeStrings 错误码字符串映射
var errCodeStrings = map[ErrCode]string{
	ErrCodeOK:            "OK",
	ErrCodeInternalError: "internal error",
}
