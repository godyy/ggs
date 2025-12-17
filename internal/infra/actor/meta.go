package actor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/godyy/gactor"
	pkgerrors "github.com/pkg/errors"
	redis "github.com/redis/go-redis/v9"
)

const metaTimeout = 5 * time.Second

// Meta Actor Meta 数据.
type Meta struct {
	UID    gactor.ActorUID // Actor唯一ID
	NodeId string          // 节点ID
}

// GetActorUID 获取 Actor 唯一ID.
func (m *Meta) GetActorUID() gactor.ActorUID {
	return m.UID
}

// GetNodeId 获取节点ID.
func (m *Meta) GetNodeId() string {
	return m.NodeId
}

// IsNodeValid 返回 Meta 的节点信息是否有效.
func (m *Meta) IsNodeValid() bool {
	return m.NodeId != ""
}

// UpdateNode 更新 Actor Meta 的节点信息.
func (m *Meta) UpdateNode(nodeId string) {
	m.NodeId = nodeId
}

// NewMeta 构造 Actor Meta 数据.
func NewMeta(uid gactor.ActorUID) *Meta {
	return &Meta{
		UID: uid,
	}
}

// NewMetaOnNode 构造固定于某一节点上的 Actor Meta 数据.
func NewMetaOnNode(uid gactor.ActorUID, nodeId string) *Meta {
	return &Meta{
		UID:    uid,
		NodeId: nodeId,
	}
}

// genMetaKey 生成 Meta 数据的 Key.
func genMetaKey(uid gactor.ActorUID) string {
	return fmt.Sprintf("actormeta:%d:%d", uid.Category, uid.ID)
}

// MetaDriver Meta 数据驱动.
type MetaDriver struct {
	redisCli redis.UniversalClient
	metas    map[uint16]map[int64]*Meta
}

func NewMetaDriver(redisCli redis.UniversalClient) *MetaDriver {
	return &MetaDriver{
		redisCli: redisCli,
	}
}

// AddActor 添加 Actor Meta 数据.
func (m *MetaDriver) AddActor(meta *Meta) error {
	metaString, err := json.Marshal(meta)
	if err != nil {
		return pkgerrors.WithMessage(err, "marshal meta")
	}

	key := genMetaKey(meta.UID)
	ctx, cancel := context.WithTimeout(context.Background(), metaTimeout)
	defer cancel()

	if err := m.redisCli.Set(ctx, key, metaString, 0).Err(); err != nil {
		return err
	}

	return nil
}

// GetActor 获取 Actor Meta 数据.
func (m *MetaDriver) GetActor(uid gactor.ActorUID) (*Meta, error) {
	key := genMetaKey(uid)
	ctx, cancel := context.WithTimeout(context.Background(), metaTimeout)
	defer cancel()

	metaString, err := m.redisCli.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, gactor.ErrMetaNotExists
		}
		return nil, err
	}

	meta := &Meta{}
	if err := json.Unmarshal([]byte(metaString), meta); err != nil {
		return nil, pkgerrors.WithMessage(err, "unmarshal meta")
	}

	return meta, nil
}

// GetMeta 获取 Actor Meta 数据.
// 当 Meta 数据不存在时返回 ErrMetaNotExists.
func (m *MetaDriver) GetMeta(uid gactor.ActorUID) (gactor.Meta, error) {
	return m.GetActor(uid)
}
