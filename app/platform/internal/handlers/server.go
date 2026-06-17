package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/platform/internal/app"
	"github.com/godyy/ggs/app/platform/internal/infra/repo"
	"github.com/godyy/ggs/app/platform/internal/models/httpproto"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/nodeutil"
	"github.com/godyy/ggs/internal/infra/actor"
	mongomodels "github.com/godyy/ggs/internal/infra/mongo/models"
	"github.com/godyy/ggs/internal/utils/ginutils"
	"github.com/godyy/ggskit/infra/cluster"
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
	nodeId := cluster.MakeNodeID(consts.NodeGame, nodeutil.MakeServerNodeName(req.ID))

	// 创建服务器.
	server := &mongomodels.Server{
		ID:     req.ID,
		Name:   req.Name,
		NodeId: nodeId,
	}
	if err := repo.Server.CreateServer(context.Background(), server); err != nil {
		return err
	}

	// 预注册服务器 Actor.
	if _, err := app.ActorRegistry().RegisterActor(gactor.ActorRegisterParams{
		UID: gactor.ActorUID{
			Category: actor.CategoryServer.ActorCategory(),
			ID:       server.ID,
		},
		NodeId:  server.NodeId,
		LeaseId: app.ActorRegistry().MakeLeaseID(),
	}); err != nil {
		return err
	}
	return nil
}
