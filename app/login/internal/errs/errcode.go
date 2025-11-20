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

func (ec ErrCode) Code() errs.ErrCode {
	return errs.ErrCode(ec)
}

func (ec ErrCode) Msg() string {
	return errCodeStrings[ec]
}

// 错误码
const (
	ErrCodeOK                          = ErrCode(0) // OK
	ErrCodeInternalError               = ErrCode(1) // 服务器内部错误
	ErrCodeAuthFailed                  = ErrCode(2) // 鉴权失败
	ErrCodeServerUnavailable           = ErrCode(3) // 服务器不可用
	ErrCodeCharacterCountLimited       = ErrCode(4) // 角色数量达到上限
	ErrCodeServerCharacterCountLimited = ErrCode(5) // 服务器角色数量达到上限
	ErrCodeCharacterNotExist           = ErrCode(6) // 角色不存在
)

// ErrCodeStrings 错误码字符串映射
var errCodeStrings = map[ErrCode]string{
	ErrCodeOK:                          "OK",
	ErrCodeInternalError:               "internal error",
	ErrCodeAuthFailed:                  "auth failed",
	ErrCodeServerUnavailable:           "server unavailable",
	ErrCodeCharacterCountLimited:       "character count limited",
	ErrCodeServerCharacterCountLimited: "server character count limited",
	ErrCodeCharacterNotExist:           "character not exist",
}
