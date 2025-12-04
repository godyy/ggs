package pprof

import (
	"fmt"
	"net/http"
	stdpprof "net/http/pprof"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterHTTP 注册 HTTP 路由，默认路径为 /debug/pprof。
func RegisterHTTP(mux *http.ServeMux, basePath string) {
	if mux == nil {
		return
	}
	base := normalizeBase(basePath)
	// Redirect base to trailing-slash path for correct relative links
	mux.HandleFunc(base, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, base+"/", http.StatusMovedPermanently)
	})
	mux.HandleFunc(base+"/", stdpprof.Index)
	mux.HandleFunc(base+"/cmdline", stdpprof.Cmdline)
	mux.HandleFunc(base+"/profile", stdpprof.Profile)
	mux.HandleFunc(base+"/symbol", stdpprof.Symbol)
	mux.HandleFunc(base+"/trace", stdpprof.Trace)
	mux.Handle(base+"/goroutine", stdpprof.Handler("goroutine"))
	mux.Handle(base+"/heap", stdpprof.Handler("heap"))
	mux.Handle(base+"/allocs", stdpprof.Handler("allocs"))
	mux.Handle(base+"/block", stdpprof.Handler("block"))
	mux.Handle(base+"/mutex", stdpprof.Handler("mutex"))
	mux.Handle(base+"/threadcreate", stdpprof.Handler("threadcreate"))
}

// RegisterGin 注册 Gin 路由，默认路径为 /debug/pprof。
func RegisterGin(engine *gin.Engine, basePath string) {
	if engine == nil {
		return
	}
	base := normalizeBase(basePath)
	// Redirect base to trailing-slash path for correct relative links
	engine.GET(base, func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, base+"/")
	})
	engine.GET(base+"/", gin.WrapF(stdpprof.Index))
	engine.GET(base+"/cmdline", gin.WrapF(stdpprof.Cmdline))
	engine.GET(base+"/profile", gin.WrapF(stdpprof.Profile))
	engine.POST(base+"/symbol", gin.WrapF(stdpprof.Symbol))
	engine.GET(base+"/symbol", gin.WrapF(stdpprof.Symbol))
	engine.GET(base+"/trace", gin.WrapF(stdpprof.Trace))
	engine.GET(base+"/goroutine", gin.WrapH(stdpprof.Handler("goroutine")))
	engine.GET(base+"/heap", gin.WrapH(stdpprof.Handler("heap")))
	engine.GET(base+"/allocs", gin.WrapH(stdpprof.Handler("allocs")))
	engine.GET(base+"/block", gin.WrapH(stdpprof.Handler("block")))
	engine.GET(base+"/mutex", gin.WrapH(stdpprof.Handler("mutex")))
	engine.GET(base+"/threadcreate", gin.WrapH(stdpprof.Handler("threadcreate")))
}

// normalizeBase 标准化路径，确保以 / 开头，不以 / 结尾。
func normalizeBase(p string) string {
	b := strings.TrimSpace(p)
	if b == "" {
		b = "/debug/pprof"
	}
	b = strings.TrimSuffix(b, "/")
	if !strings.HasPrefix(b, "/") {
		b = "/" + b
	}
	return b
}

// Path 拼接路径，确保以 / 开头，不以 / 结尾。
func Path(basePath string, sub string) string {
	base := normalizeBase(basePath)
	sub = strings.TrimPrefix(sub, "/")
	return fmt.Sprintf("%s/%s", base, sub)
}
