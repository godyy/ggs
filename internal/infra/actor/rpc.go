package actor

import (
	pbcommon "github.com/godyy/ggs/internal/infra/actor/protocol/pb/common"
	"google.golang.org/protobuf/proto"
)

// RPCResult RPC调用结果.
type RPCResult struct {
	reply proto.Message
	err   error
}

func NewRPCResult(reply proto.Message, err error) RPCResult {
	return RPCResult{reply: reply, err: err}
}

// RPCResult 是否调用成功.
func (r *RPCResult) Success() bool {
	if r.err != nil {
		return false
	}
	if _, ok := r.reply.(*pbcommon.Error); ok {
		return false
	}
	return true
}

// Reply 获取调用返回值.
func (r *RPCResult) Reply() proto.Message {
	return r.reply
}

// RPCResult 返回错误.
func (r *RPCResult) Err() error {
	if r.err != nil {
		return r.err
	}
	if rErr, ok := r.reply.(*pbcommon.Error); ok {
		return &PbError{Err: rErr}
	}
	return nil
}

// ErrCode 获取错误码.
func (r *RPCResult) ErrCode() pbcommon.ErrCode {
	if r.reply == nil {
		return -1
	}
	if rErr, ok := r.reply.(*pbcommon.Error); ok {
		return rErr.Code
	}
	return -1
}
