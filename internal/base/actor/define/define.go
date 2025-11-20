package define

import (
	"github.com/godyy/gactor"
)

var (
	defineList []gactor.IActorDefine
)

func RegisterDefine(define ...gactor.IActorDefine) {
	defineList = append(defineList, define...)
}

func GetDefineList() []gactor.IActorDefine {
	return defineList
}
