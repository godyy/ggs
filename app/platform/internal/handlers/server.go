package handlers

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/godyy/gactor"
	"github.com/godyy/ggs/app/platform/internal/app"
	"github.com/godyy/ggs/app/platform/internal/infra/repo"
	"github.com/godyy/ggs/app/platform/internal/models/httpproto"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/base/nodeutil"
	"github.com/godyy/ggs/internal/infra/actor"
	pbs2s "github.com/godyy/ggs/internal/infra/actor/protocol/pb/s2s"
	mongomodels "github.com/godyy/ggs/internal/infra/mongo/models"
	"github.com/godyy/ggs/internal/utils/ginutils"
	"github.com/godyy/ggskit/infra/cluster"
	"github.com/pkg/errors"
	pkgerrors "github.com/pkg/errors"
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
		group.POST("/reload-gdconf", ginutils.WrapHandlerJsonJson(s.handleServerReloadGDConf))
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

func (s *serverHandler) handleServerReloadGDConf(c *gin.Context, req *httpproto.ServerReloadGDConfReq, resp *httpproto.ServerReloadGDConfResp) error {
	serverIds := req.ServerIds

	if req.AllServers {
		var err error
		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		serverIds, err = repo.Server.GetAllServerIds(ctx)
		if err != nil {
			cancel()
			return pkgerrors.WithMessage(err, "get all server ids failed")
		}
		cancel()
	}

	if len(serverIds) == 0 {
		return errors.New("no servers")
	}

	if !req.All && len(req.Tables) == 0 {
		return errors.New("no tables")
	}

	const batchLimit = 100
	const reloadTimeout = 10 * time.Second

	i := 0
	for i < len(serverIds) {
		j := i + batchLimit
		if j > len(serverIds) {
			j = len(serverIds)
		}
		wg := sync.WaitGroup{}
		chFailed := make(chan httpproto.ServerReloadFailed, batchLimit)
		for i < j {
			wg.Add(1)
			go func(serverId int64) {
				defer wg.Done()
				msg, err := app.ActorService().RPCWithTimeout(gactor.ActorUID{
					Category: actor.CategoryServer.ActorCategory(),
					ID:       serverId,
				}, &pbs2s.ReloadGDConfReq{
					All:    req.All,
					Tables: req.Tables,
				}, reloadTimeout)
				if err != nil {
					chFailed <- httpproto.ServerReloadFailed{
						ServerId: serverId,
						Error:    err.Error(),
					}
					return
				}
				resp := msg.(*pbs2s.ReloadGDConfResp)
				if !resp.Success {
					chFailed <- httpproto.ServerReloadFailed{
						ServerId: serverId,
						Error:    resp.Err,
					}
					return
				}
			}(serverIds[i])
			i++
		}
		wg.Wait()
		for len(chFailed) > 0 {
			resp.ServerFailed = append(resp.ServerFailed, <-chFailed)
		}
	}

	resp.Success = len(resp.ServerFailed) == 0

	return nil
}
