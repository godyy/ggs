package errs

import (
	"fmt"

	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/proto/pb/common"
)

// PbError 将Error协议结构封装实现error.
type PbError struct {
	Err *pbcommon.Error
}

func (e PbError) Error() string {
	return fmt.Sprintf("{%+v}", e.Err)
}

// WithPbError 创建一个PbError.
func WithPbError(code int32, args ...*pbcommon.ErrArg) *PbError {
	return &PbError{
		Err: &pbcommon.Error{
			Code: code,
			Args: args,
		},
	}
}

// WithC2SPbError 创建一个C2S PbError.
func WithC2SPbError(code pbc2s.ErrCode, args ...*pbcommon.ErrArg) *PbError {
	return WithPbError(int32(code), args...)
}
