package actor

import (
	"github.com/godyy/gactor"
	model1 "github.com/godyy/ggs/internal/infra/actor/model"
	"github.com/godyy/ggskit/infra/actor"
)

type ActorUID = actor.ActorUID

type Actor = actor.Actor

type CActor = actor.CActor

type TimerId = gactor.TimerId

func GetActorModule[M actor.Module](a actor.ActorWithModule, autoCreate bool) M {
	return actor.GetActorModule[M](a, autoCreate)
}

// model 模型接口.
// 用于限制ActorWithModel的模型类型.
type model interface {
	comparable
	model1.ModelDirty
}

// modelWithModule 模型接口.
// 用于限制ActorWithModule的模型类型.
type modelWithModule interface {
	comparable
	model1.ModelWithModule
}

// ActorWithModel 携带数据模型的Actor基础封装.
type ActorWithModel[Model model] struct {
	ActorSugared
	persistor
	Model Model
}

func NewActorWithModel[Model model](actor actor.Actor) ActorWithModel[Model] {
	return ActorWithModel[Model]{ActorSugared: ActorSugared{Actor: actor}}
}

func (a *ActorWithModel[Model]) GetModel() actor.Model {
	return a.Model
}

func (a *ActorWithModel[Model]) OnModelDirty() {
	if ok, _ := a.Model.IsDirty(); !ok {
		return
	}
	DelaySave(a, actorSaveDelay)
}

func (a *ActorWithModel[Model]) OnStart() error {
	// 加载model数据
	exists, err := LoadModel(a)
	if err != nil {
		return err
	}

	// 若数据不存在. 准备存储新数据.
	if !exists {
		a.Model.SetAllDirty()
		a.OnModelDirty()
	}

	return nil
}

func (a *ActorWithModel[Model]) OnStop() error {
	// 持久化脏数据.
	if ok, _ := a.Model.IsDirty(); ok {
		if err := SaveModel(a); err != nil {
			return err
		}
	}

	// 释放model
	var zeroModel Model
	if zeroModel != a.Model {
		a.Model.Release()
	}

	return nil
}

// ActorWithModule 携带模块的Actor基础封装.
type ActorWithModule[Model modelWithModule] struct {
	ActorSugared
	persistor
	Model Model
}

func NewActorWithModule[Model modelWithModule](actor actor.Actor) ActorWithModule[Model] {
	return ActorWithModule[Model]{ActorSugared: ActorSugared{Actor: actor}}
}

func (a *ActorWithModule[Model]) GetModel() actor.Model {
	return a.Model
}

func (a *ActorWithModule[Model]) GetModuleWithModule() actor.ModelWithModule {
	return a.Model
}

func (a *ActorWithModule[Model]) OnModelDirty() {
	if ok, _ := a.Model.IsDirty(); !ok {
		return
	}
	DelaySave(a, actorSaveDelay)
}

func (a *ActorWithModule[Model]) SetDirtyModules(mk ...actor.ModuleKey) {
	if len(mk) == 0 {
		return
	}
	for _, m := range mk {
		a.Model.SetDirtyModule(m)
	}
	a.OnModelDirty()
}

func (a *ActorWithModule[Model]) SetAllDirty() {
	a.Model.SetAllDirty()
	a.OnModelDirty()
}

func (a *ActorWithModule[Model]) OnStart() error {
	// 加载model数据
	exists, err := LoadModel(a)
	if err != nil {
		return err
	}

	// 若数据不存在. 准备存储新数据.
	if !exists {
		a.Model.SetAllDirty()
		a.OnModelDirty()
	}

	return nil
}

func (a *ActorWithModule[Model]) OnStop() error {
	// 持久化脏数据.
	if ok, _ := a.Model.IsDirty(); ok {
		if err := SaveModel(a); err != nil {
			return err
		}
	}

	// 释放model
	var zeroModel Model
	if zeroModel != a.Model {
		a.Model.Release()
	}

	return nil
}

// CActorWithModel 携带数据模型的CActor基础封装.
type CActorWithModel[Model model] struct {
	CActorSugared
	persistor
	Model Model
}

func NewCActorWithModel[Model model](actor actor.CActor) CActorWithModel[Model] {
	return CActorWithModel[Model]{CActorSugared: CActorSugared{CActor: actor}}
}

func (a *CActorWithModel[Model]) GetModel() actor.Model {
	return a.Model
}

func (a *CActorWithModel[Model]) OnModelDirty() {
	if ok, _ := a.Model.IsDirty(); !ok {
		return
	}
	DelaySave(a, actorSaveDelay)
}

func (a *CActorWithModel[Model]) OnStart() error {
	// 加载model数据
	exists, err := LoadModel(a)
	if err != nil {
		return err
	}

	// 若数据不存在. 准备存储新数据.
	if !exists {
		a.Model.SetAllDirty()
		a.OnModelDirty()
	}

	return nil
}

func (a *CActorWithModel[Model]) OnStop() error {
	// 持久化脏数据.
	if ok, _ := a.Model.IsDirty(); ok {
		if err := SaveModel(a); err != nil {
			return err
		}
	}

	// 释放model
	var zeroModel Model
	if zeroModel != a.Model {
		a.Model.Release()
	}

	return nil
}

// CActorWithModule 携带模块的CActor基础封装.
type CActorWithModule[Model modelWithModule] struct {
	CActorSugared
	persistor
	Model Model
}

func NewCActorWithModule[Model modelWithModule](actor actor.CActor) CActorWithModule[Model] {
	return CActorWithModule[Model]{CActorSugared: CActorSugared{CActor: actor}}
}

func (a *CActorWithModule[Model]) GetModel() actor.Model {
	return a.Model
}

func (a *CActorWithModule[Model]) GetModelWithModule() actor.ModelWithModule {
	return a.Model
}

func (a *CActorWithModule[Model]) OnModelDirty() {
	if ok, _ := a.Model.IsDirty(); !ok {
		return
	}
	DelaySave(a, actorSaveDelay)
}

func (a *CActorWithModule[Model]) SetDirtyModules(mk ...actor.ModuleKey) {
	if len(mk) == 0 {
		return
	}
	for _, m := range mk {
		a.Model.SetDirtyModule(m)
	}
	a.OnModelDirty()
}

func (a *CActorWithModule[Model]) SetAllDirty() {
	a.Model.SetAllDirty()
	a.OnModelDirty()
}

func (a *CActorWithModule[Model]) OnStart() error {
	// 加载model数据
	exists, err := LoadModel(a)
	if err != nil {
		return err
	}

	// 若数据不存在. 准备存储新数据.
	if !exists {
		a.Model.SetAllDirty()
		a.OnModelDirty()
	}

	return nil
}

func (a *CActorWithModule[Model]) OnStop() error {
	// 持久化脏数据.
	if ok, _ := a.Model.IsDirty(); ok {
		if err := SaveModel(a); err != nil {
			return err
		}
	}

	// 释放model
	var zeroModel Model
	if zeroModel != a.Model {
		a.Model.Release()
	}

	return nil
}
