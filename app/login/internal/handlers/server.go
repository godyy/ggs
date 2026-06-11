package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/godyy/ggs/app/login/httpproto"
	"github.com/godyy/ggs/app/login/internal/base/consts"
	"github.com/godyy/ggs/app/login/internal/infra/repo"
	cginutils "github.com/godyy/ggs/internal/utils/ginutils"
	"github.com/godyy/ggskit/utils/ctxutils"
)

type serverHandler struct{}

func init() {
	reigsterHandler(&serverHandler{})
}

// groupPath 返回路由组路径.
func (s *serverHandler) groupPath() string {
	return "/server"
}

// setupRoutes 配置路由.
func (s *serverHandler) setupRoutes(root *gin.RouterGroup, version string) {
	group := root.Group(s.groupPath())
	{
		group.GET("/list", cginutils.WrapHandlerQueryJson(s.handleServerList))
	}
}

func (h *serverHandler) handleServerList(c *gin.Context, req *httpproto.GetServerListReq, resp *httpproto.GetServerListResp) error {
	ctx, cancel := ctxutils.WithTimeout(context.Background(), consts.DefaultTimeout)
	defer cancel()

	allServers, err := repo.Server.GetAllServers(ctx)
	if err != nil {
		return err
	}

	resp.ServerList = make([]httpproto.ServerInfo, len(allServers))
	for i, server := range allServers {
		resp.ServerList[i] = httpproto.ServerInfo{
			ID:   server.ID,
			Name: server.Name,
		}
	}

	return nil
}
