package actors

import (
	"fmt"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/infra/actor"
)

var (
	defineList []gactor.ActorDefine
	defineMap  = map[actor.Category]gactor.ActorDefine{}
)

// registerDefine 注册Actor定义.
func registerDefine(define gactor.ActorDefine) {
	if actor.Category(define.Category()).String() == "" {
		panic(fmt.Sprintf("category %d not exists", define.Category()))
	}

	if _, ok := defineMap[actor.Category(define.Category())]; ok {
		panic(fmt.Sprintf("category %d already registered: %v", define.Category(), define))
	}

	defineList = append(defineList, define)
	defineMap[actor.Category(define.Category())] = define
}

// GetDefineList 获取Actor定义列表.
func GetDefineList() []gactor.ActorDefine {
	return defineList
}
