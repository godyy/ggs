package ginutils

import "github.com/gin-gonic/gin"

// WrapHandlerJsonJson json格式的请求/响应体.
func WrapHandlerJsonJson[Req, Resp any](h func(c *gin.Context, req *Req, resp *Resp) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequest(c, err)
			return
		}

		var resp Resp
		if err := h(c, &req, &resp); err != nil {
			FailWithError(c, err)
			return
		}

		OK(c, &resp)
	}
}

// WrapQueryJson 包装请求数据为 query 参数, 响应数据为 json 的处理器.
func WrapHandlerQueryJson[Req, Resp any](h func(c *gin.Context, req *Req, resp *Resp) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindQuery(&req); err != nil {
			BadRequest(c, err)
			return
		}

		var resp Resp
		if err := h(c, &req, &resp); err != nil {
			FailWithError(c, err)
			return
		}

		OK(c, &resp)
	}
}

// WrapHandlerJsonNone json格式的请求, 无响应体.
func WrapHandlerJsonNone[Req any](h func(c *gin.Context, req *Req) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequest(c, err)
			return
		}

		if err := h(c, &req); err != nil {
			FailWithError(c, err)
			return
		}

		OK(c, nil)
	}
}
