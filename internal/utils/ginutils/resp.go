package ginutils

import (
	"net/http"

	"github.com/godyy/ggs/internal/errs"

	"github.com/gin-gonic/gin"
)

// commonResp 通用响应
type commonResp struct {
	Code int    `json:"code"`           // 响应代码.
	Msg  string `json:"msg"`            // 响应消息.
	Data any    `json:"data,omitempty"` // 响应数据.
}

// AbortWithStatusError 终止请求, 返回指定的状态码以及携带错误信息的json响应体.
func AbortWithStatusError(c *gin.Context, statusCode int, err error) {
	var (
		code int
		msg  string
	)
	switch e := err.(type) {
	case errs.Error:
		code = int(e.Code())
		msg = e.Msg()
	default:
		code = -1
		msg = err.Error()
	}
	c.AbortWithStatusJSON(statusCode, commonResp{
		Code: code,
		Msg:  msg,
	})
}

// AbortWithError 终止请求, 返回默认的状态码(http.StatusOK)以及携带错误信息的json响应体.
func AbortWithError(c *gin.Context, err error) {
	AbortWithStatusError(c, http.StatusOK, err)
}

// BadRequest 错误的请求参数.
func BadRequest(c *gin.Context, err error) {
	AbortWithStatusError(c, http.StatusBadRequest, err)
}

// OK
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, commonResp{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

// FailWithError 失败的响应.
func FailWithError(c *gin.Context, err error) {
	AbortWithStatusError(c, http.StatusOK, err)
}
