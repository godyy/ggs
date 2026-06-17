package model

import (
	"errors"

	"github.com/godyy/ggskit/infra/actor"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ModelDirty 脏数据模型.
type ModelDirty interface {
	actor.ModelDirty
	// SetAllDirty 设置全脏位.
	SetAllDirty()
}

// ModelWithModule 脏数据模型，携带模块.
type ModelWithModule interface {
	actor.ModelWithModule

	// SetDirtyModule 设置脏模块.
	SetDirtyModule(mk actor.ModuleKey)

	// SetAllDirty 设置全脏位.
	SetAllDirty()
}

// Dirty 脏位.
type Dirty bool

// SetAllDirty 设置全脏位.
func (d *Dirty) SetAllDirty() {
	*d = true
}

// IsDirty 是否有脏数据.
func (d *Dirty) IsDirty() (bool, bool) {
	return bool(*d), bool(*d)
}

// ClearDirty 清除脏数据.
func (d *Dirty) ClearDirty() {
	*d = false
}

// MarshalBSONDirty 序列化脏数据.
func (d *Dirty) MarshalBSONDirty() ([]byte, error) {
	return nil, errors.New("should not be invoked")
}

// Dirties 脏数据搜集器.
type Dirties struct {
	dirties bson.M // 脏数据
	all     bool   // 是否全脏位.
}

// NewDirties 构造脏数据模型.
func NewDirties() Dirties {
	return Dirties{dirties: make(bson.M)}
}

// SetDirty 设置脏数据.
func (d *Dirties) SetDirty(key string, value any) {
	d.dirties[key] = value
}

// SetDirtyAll 设置全脏位.
func (d *Dirties) SetAllDirty() {
	d.all = true
}

// IsDirty 是否有脏数据.
func (d *Dirties) IsDirty() (dirty bool, all bool) {
	all = d.all
	dirty = all || len(d.dirties) > 0
	return
}

// ClearDirty 清除脏数据.
func (d *Dirties) ClearDirty() {
	d.dirties = make(bson.M)
	d.all = false
}

// Release 释放资源.
func (d *Dirties) Release() {
	d.dirties = nil
}

// MarshalBSONDirty 序列化脏数据.
func (d *Dirties) MarshalBSONDirty() ([]byte, error) {
	return bson.Marshal(d.dirties)
}
