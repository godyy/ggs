package gdconf

import (
	"sync/atomic"
)

var levelOfItems atomic.Value

func init() {
	RegisterAfterLoadFunc(TblNameItem, func() error {
		v := make(map[int32][]*Item)
		for _, item := range TblItem().All() {
			v[item.Level] = append(v[item.Level], item)
		}
		levelOfItems.Store(v)
		return nil
	}, 0)
}

func GetLevelOfItems(level int32) []*Item {
	return levelOfItems.Load().(map[int32][]*Item)[level]
}
