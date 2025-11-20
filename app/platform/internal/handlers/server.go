package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/platform/internal/app"
	"github.com/godyy/ggs/app/platform/internal/base/models/httpproto"
	"github.com/godyy/ggs/app/platform/internal/data/repository"
	"github.com/godyy/ggs/internal/base/actor"
	dbmodels "github.com/godyy/ggs/internal/base/models/db"
	"github.com/godyy/ggs/internal/utils/ginutils"
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
		group.POST("/create", ginutils.WrapHandlerJsonNone(s.handleServerCreate))
	}
}

func (s *serverHandler) handleServerCreate(c *gin.Context, req *httpproto.ServerCreateReq) error {
	// 创建服务器.
	server := &dbmodels.Server{
		ID:     req.ID,
		Name:   req.Name,
		NodeId: req.NodeId,
	}
	if err := repository.Server.CreateServer(context.Background(), server); err != nil {
		return err
	}

	// 创建ActorMeta信息.
	serverActorMeta := &gactor.Meta{
		Category:   actor.CategoryServer.Uint16(),
		ID:         server.ID,
		Deployment: gactor.NewDeploymentOnNode(server.NodeId),
	}
	if err := app.ActorMetaDriver().AddMeta(serverActorMeta); err != nil {
		return err
	}
	return nil
}
