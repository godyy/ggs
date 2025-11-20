package errs

import (
	"github.com/godyy/ggs/internal/errs"
)

// WithErrCodeMsg 创建包含错误码及错误信息的 Error
func WithErrCodeMsg(code ErrCode, msg string) errs.Error {
	return errs.WithErrCodeMsg(code.Code(), msg)
}

// WithErrCodeMsgf 创建包含错误码及格式化错误信息的 Error
func WithErrCodeMsgf(code ErrCode, format string, a ...any) errs.Error {
	return errs.WithErrCodeMsgf(code.Code(), format, a...)
}

// WithErrCodeErr 创建包含错误码及错误的 Error
func WithErrCodeErr(code ErrCode, err error) errs.Error {
	return errs.WithErrCodeErr(code.Code(), err)
}

// InternalErrorWithMsg 创建错误码ErrCodeInternalError的错误.
func InternalErrorWithMsg(msg string) errs.Error {
	return errs.WithErrCodeMsg(ErrCodeInternalError.Code(), msg)
}

// InernalErrorWithErr 创建错误码ErrCodeInternalError的错误.
func InernalErrorWithErr(err error) errs.Error {
	return errs.WithErrCodeErr(ErrCodeInternalError.Code(), err)
}
