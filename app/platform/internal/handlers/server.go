package handlers

import (
	"context"

	iodb "github.com/godyy/ggs/internal/core/db/io"
	mdb "github.com/godyy/ggs/internal/core/db/models"

	"github.com/gin-gonic/gin"
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/platform/internal/models/httpproto"
	mactor "github.com/godyy/ggs/app/platform/internal/modules/actor"
	"github.com/godyy/ggs/internal/core/actor"
	libmongo "github.com/godyy/ggs/internal/libs/db/mongo"
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
	server := &mdb.Server{
		ID:     req.ID,
		Name:   req.Name,
		NodeId: req.NodeId,
	}
	if err := iodb.Server.CreateServer(context.Background(), libmongo.Inst(), server); err != nil {
		return err
	}

	// 创建ActorMeta信息.
	serverActorMeta := &gactor.Meta{
		Category:   actor.CategoryServer,
		ID:         server.ID,
		Deployment: gactor.NewDeploymentOnNode(server.NodeId),
	}
	if err := mactor.AddMeta(serverActorMeta); err != nil {
		return err
	}
	return nil
}
