package actor

import (
	"fmt"

	pbcommon "github.com/godyy/ggs/internal/infra/actor/protocol/pb/common"
)

// PbError 将Error协议结构封装实现error.
type PbError struct {
	Err *pbcommon.Error
}

func (e PbError) Error() string {
	return fmt.Sprintf("{%+v}", e.Err)
}

// WithPbError 创建一个PbError.
func WithPbError(code pbcommon.ErrCode, args ...*pbcommon.ErrArg) *PbError {
	return &PbError{
		Err: &pbcommon.Error{
			Code: code,
			Args: args,
		},
	}
}
