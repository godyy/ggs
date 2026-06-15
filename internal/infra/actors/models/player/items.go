package player

// Item 道具信息.
// 表示某个道具ID及其持有数量.
type Item struct {
	ID  int32 `bson:"id"`
	Num int64 `bson:"num"`
}

// Items 道具模块.
// 用于集中维护玩家道具数据，并提供获取/增减等操作方法.
type Items struct {
	moduleBase[*Items]
	Items map[int32]int64
}

// ModuleKey 模块关键字.
func (m *Items) ModuleKey() string {
	return "items"
}

// GetNum 获取指定道具的持有数量.
func (m *Items) GetNum(id int32) int64 {
	if id == 0 {
		return 0
	}
	if m.Items == nil {
		return 0
	}
	return m.Items[id]
}

// GetItem 获取指定道具信息.
// 若不存在返回 (Item{}, false).
func (m *Items) GetItem(id int32) (Item, bool) {
	if id == 0 {
		return Item{}, false
	}
	if m.Items == nil {
		return Item{}, false
	}
	num, ok := m.Items[id]
	if !ok {
		return Item{}, false
	}
	return Item{ID: id, Num: num}, true
}

// Add 增加指定道具数量.
// 返回增加后的道具数量；若 id==0 或 num<=0 则不变更数据.
func (m *Items) Add(id int32, num int64) int64 {
	if id == 0 || num <= 0 {
		return m.GetNum(id)
	}

	if m.Items == nil {
		m.Items = make(map[int32]int64, 8)
	}

	after := m.Items[id] + num
	if after <= 0 {
		delete(m.Items, id)
		m.SetDirty()
		return 0
	}

	m.Items[id] = after
	m.SetDirty()
	return after
}

// Sub 扣除指定道具数量.
// ok==false 表示道具不存在或数量不足；after 返回扣除后的数量（或当前数量）.
func (m *Items) Sub(id int32, num int64) (after int64, ok bool) {
	if id == 0 {
		return 0, false
	}
	if num <= 0 {
		return m.GetNum(id), true
	}

	if m.Items == nil {
		return 0, false
	}

	cur, exists := m.Items[id]
	if !exists || cur < num {
		if !exists {
			return 0, false
		}
		return cur, false
	}

	after = cur - num
	if after == 0 {
		delete(m.Items, id)
		m.SetDirty()
		return 0, true
	}
	m.Items[id] = after
	m.SetDirty()
	return after, true
}
