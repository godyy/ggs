package errs

import (
	"testing"

	pbc2s "github.com/godyy/ggs/internal/proto/pb/c2s"
	pbcommon "github.com/godyy/ggs/internal/proto/pb/common"
)

func TestPbError_Error(t *testing.T) {
	err := &PbError{
		Err: &pbcommon.Error{
			Code: int32(pbc2s.ErrCode_ECLoginTimeout),
			Args: []*pbcommon.ErrArg{&pbcommon.ErrArg{Value: &pbcommon.ErrArg_S{S: "123"}}},
		},
	}
	t.Log(err.Error())
}
