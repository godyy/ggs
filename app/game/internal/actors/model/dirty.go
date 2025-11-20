package model

import "go.mongodb.org/mongo-driver/v2/bson"

// DirtyModel 脏数据模型.
type DirtyModel struct {
	actor   Actor  // 关联Actor.
	dirties bson.M // 脏数据
	all     bool   // 是否全脏
}

// NewDirtyModel 构造脏数据模型.
func NewDirtyModel(actor Actor) *DirtyModel {
	return &DirtyModel{actor: actor}
}

// SetDirty 设置脏数据.
func (dm *DirtyModel) SetDirty(key string, value any) {
	if dm.dirties == nil {
		dm.dirties = make(bson.M)
	}
	dm.dirties[key] = value
	dm.actor.OnModelDirty()
}

// SetDirtyAll 设置全脏位.
func (dm *DirtyModel) SetDirtyAll() {
	dm.all = true
	dm.actor.OnModelDirty()
}

// IsDirty 是否有脏数据.
func (dm *DirtyModel) IsDirty() (dirty bool, all bool) {
	all = dm.all
	dirty = all || len(dm.dirties) > 0
	return
}

// ClearDirty 清除脏数据.
func (dm *DirtyModel) ClearDirty() {
	dm.dirties = nil
	dm.all = false
}

// MarshalBSONDirty 序列化脏数据.
func (dm *DirtyModel) MarshalBSONDirty() ([]byte, error) {
	return bson.Marshal(dm.dirties)
}

func (dm *DirtyModel) Release() {
	dm.actor = nil
	dm.dirties = nil
}
