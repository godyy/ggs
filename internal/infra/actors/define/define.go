package define

import (
	"github.com/godyy/gactor"
)

var (
	defineList []gactor.ActorDefine
)

func RegisterDefine(define ...gactor.ActorDefine) {
	defineList = append(defineList, define...)
}

func GetDefineList() []gactor.ActorDefine {
	return defineList
}
