package chat

import "github.com/godyy/ggs/internal/infra/actor/model"

// Mgr Mgr 数据模型.
type Mgr struct {
	model.Dirty               `bson:"-"`       // 集成脏标记位
	model.ModelWithID[string] `bson:",inline"` // 集成具备ID字段的模型

	RoomIDIncr int64 `bson:"roomIDIncr"` // 房间ID递增器.
}

func NewMgr() *Mgr {
	return &Mgr{
		ModelWithID: model.NewModuleWithID("chatmgr"),
	}
}

func (m *Mgr) GetCollection() string {
	return model.CollSingleton
}

// NextRoomID 获取下一个房间ID.
func (m *Mgr) NextRoomID() int64 {
	defer m.SetAllDirty()
	m.RoomIDIncr++
	return m.RoomIDIncr
}
