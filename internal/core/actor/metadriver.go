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

// MetaDriver Meta 数据驱动.
type MetaDriver struct {
	redisCli redis.UniversalClient
	metas    map[uint16]map[int64]*gactor.Meta
}

func NewMetaDriver(redisCli redis.UniversalClient) *MetaDriver {
	return &MetaDriver{
		redisCli: redisCli,
	}
}

// genMetaKey 生成 Meta 数据的 Key.
func (m *MetaDriver) genMetaKey(category uint16, id int64) string {
	return fmt.Sprintf("actormeta:%d:%d", category, id)
}

// AddMeta 添加 Actor Meta 数据.
func (m *MetaDriver) AddMeta(meta *gactor.Meta) error {
	metaString, err := json.Marshal(meta)
	if err != nil {
		return pkgerrors.WithMessage(err, "marshal meta")
	}

	key := m.genMetaKey(meta.Category, meta.ID)
	ctx, cancel := context.WithTimeout(context.Background(), metaTimeout)
	defer cancel()

	if err := m.redisCli.Set(ctx, key, metaString, 0).Err(); err != nil {
		return err
	}

	return nil
}

// GetMeta 获取 Actor Meta 数据.
// 当 Meta 数据不存在时返回 ErrMetaNotExists.
func (m *MetaDriver) GetMeta(uid gactor.ActorUID) (*gactor.Meta, error) {
	key := m.genMetaKey(uid.Category, uid.ID)
	ctx, cancel := context.WithTimeout(context.Background(), metaTimeout)
	defer cancel()

	metaString, err := m.redisCli.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, gactor.ErrMetaNotExists
		}
		return nil, err
	}

	meta := &gactor.Meta{}
	if err := json.Unmarshal([]byte(metaString), meta); err != nil {
		return nil, pkgerrors.WithMessage(err, "unmarshal meta")
	}

	return meta, nil
}
