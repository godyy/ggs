package mode

import "fmt"

type Mode interface {
	Start()
	Stop()
}

type ModeCreator func() Mode

var modeCreators = make(map[string]ModeCreator)

func RegisterMode(mode string, creator ModeCreator) {
	if _, ok := modeCreators[mode]; ok {
		panic(fmt.Errorf("mode %s already registered", mode))
	}
	modeCreators[mode] = creator
}

func CreateMode(mode string) Mode {
	creator, ok := modeCreators[mode]
	if !ok {
		panic(fmt.Errorf("mode %s not registered", mode))
	}
	return creator()
}
