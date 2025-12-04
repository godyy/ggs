package probe

import "net/http"

// RegisterHTTP 注册 net/http 路由，默认路径为 /healthz 与 /readyz。
func RegisterHTTP(mux *http.ServeMux, healthPath, readyPath string) {
	if mux == nil {
		return
	}
	if healthPath == "" {
		healthPath = "/healthz"
	}
	if readyPath == "" {
		readyPath = "/readyz"
	}
	mux.Handle(healthPath, HealthzHandler())
	mux.Handle(readyPath, ReadyzHandler())
}
