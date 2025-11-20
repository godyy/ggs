package model

import (
	"reflect"

	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Module 数据模块
type Module interface {
	// Key 关键字,用于持久化模块的唯一标识.
	Key() string

	// SetContainer 设置模块归属容器.
	SetContainer(ModuleContainer)
}

// ModuleKey 模块关键字接口
type ModuleKey interface {
	Key() string
}

// ModelWithModules 携带模块数据的数据模型
type ModelWithModules interface {
	Model

	// ModuleRegistry 获取模块注册表.
	ModuleRegistry() *ModuleRegistry

	// SetModuleDirty 设置模块脏位.
	SetModuleDirty(m Module)
}

// ModuleGetter 模块获取器接口
type ModuleGetter interface {
	// GetModule 获取模块实例.
	GetModule(key string, autoCreate bool) Module
}

// ModuleContainer 模块容器接口
type ModuleContainer interface {
	ModuleGetter

	// OnModuleDirty 模块脏位回调.
	OnModuleDirty(key string)
}

// ModuleBase 模块基础实现.
// 集成模块需要具备的简单基础功能.
type ModuleBase[M Module] struct {
	mc ModuleContainer // 模块归属容器.
}

func NewModuleBase[M Module](mc ModuleContainer) ModuleBase[M] {
	return ModuleBase[M]{
		mc: mc,
	}
}

// SetContainer 设置模块归属容器.
func (m *ModuleBase[M]) SetContainer(mc ModuleContainer) {
	m.mc = mc
}

// SetDirty 设置模块脏位.
func (m *ModuleBase[M]) SetDirty() {
	var mm M
	m.mc.OnModuleDirty(mm.Key())
}

// ModuleSingle 单值模块.
// 用于存储单值模块数据.
type ModuleSingle[V any, Key ModuleKey] struct {
	ModuleBase[*ModuleSingle[V, Key]]
	value V
}

func (m *ModuleSingle[V, Key]) Key() string {
	var k Key
	return k.Key()
}

func (m *ModuleSingle[V, Key]) Get() V {
	return m.value
}

func (m *ModuleSingle[V, Key]) Set(v V) {
	m.value = v
	m.SetDirty()
}

func (m *ModuleSingle[V, Key]) MarshalBSONValue() (byte, []byte, error) {
	t, bytes, err := bson.MarshalValue(m.value)
	return byte(t), bytes, err
}

func (m *ModuleSingle[V, Key]) UnmarshalBSONValue(t byte, data []byte) error {
	return bson.UnmarshalValue(bson.Type(t), data, &m.value)
}

// ModuleMgr 模块管理器.
type ModuleMgr struct {
	model   ModelWithModules  // 数据模型
	modules map[string]Module // 模块实例映射表
}

func NewModuleMgr(m ModelWithModules) *ModuleMgr {
	return &ModuleMgr{
		model:   m,
		modules: make(map[string]Module, m.ModuleRegistry().Len()),
	}
}

// InitAllModules 初始化说有模块实例
func (mm *ModuleMgr) InitAllModules() {
	moduleRegistry := mm.model.ModuleRegistry()
	for _, mi := range moduleRegistry.moduleList {
		m := mi.create()
		m.SetContainer(mm)
		mm.modules[mi.key] = m
	}
}

// GetModule 获取模块实例
func (mm *ModuleMgr) GetModule(key string, autoCreate bool) Module {
	m := mm.modules[key]
	if m == nil && autoCreate {
		m = mm.model.ModuleRegistry().Create(key)
		m.SetContainer(mm)
		mm.modules[key] = m
	}

	return m
}

// OnModuleDirty 模块脏位回调.
func (mm *ModuleMgr) OnModuleDirty(key string) {
	if m := mm.modules[key]; m != nil {
		mm.model.SetModuleDirty(m)
	}
}

// Release 清理所有模块实例, 解除引用model.
// 释放时调用.
func (mm *ModuleMgr) Release() {
	for _, m := range mm.modules {
		m.SetContainer(nil)
	}
	mm.modules = nil
	mm.model = nil
}

// MarshalBSON 序列化模块实例BSON
func (mm *ModuleMgr) MarshalBSON() ([]byte, error) {
	moduleRegistry := mm.model.ModuleRegistry()
	elements := make(bson.D, 0, len(mm.modules))
	for _, mi := range moduleRegistry.moduleList {
		module := mm.modules[mi.key]
		if module == nil {
			continue
		}
		elements = append(elements, bson.E{Key: mi.key, Value: module})
	}
	return bson.Marshal(elements)
}

// UnmarshalBSON 反序列化模块实例BSON
func (mm *ModuleMgr) UnmarshalBSON(data []byte) error {
	raw := bson.Raw(data)
	moduleRegistry := mm.model.ModuleRegistry()
	for _, mi := range moduleRegistry.moduleList {
		value := raw.Lookup(mi.key)
		if value.IsZero() {
			continue
		}
		m := mi.create()
		m.SetContainer(mm)
		// if err := value.Unmarshal(m); err != nil {
		// 	return pkgerrors.WithMessagef(err, "unmarshal module %s", mi.key)
		// }
		if err := bson.UnmarshalValue(value.Type, value.Value, m); err != nil {
			return pkgerrors.WithMessagef(err, "unmarshal module %s", mi.key)
		}
		mm.modules[mi.key] = m
	}
	return nil
}

// GetModule 获取模块实例的泛型封装.
func GetModule[M Module](mg ModuleGetter, autoCreate bool) (m M) {
	module := mg.GetModule(m.Key(), autoCreate)
	if module == nil {
		panic("module " + m.Key() + " not exists")
	}
	moduleM, ok := module.(M)
	if !ok {
		panic("module " + m.Key() + " type is " + reflect.TypeOf(moduleM).Name())
	}
	return moduleM
}

// moduleInfo 模块信息
type moduleInfo struct {
	key string       // key
	typ reflect.Type // typ
}

func (mi *moduleInfo) create() Module {
	return reflect.New(mi.typ).Interface().(Module)
}

// Module 模块注册表
type ModuleRegistry struct {
	moduleList []*moduleInfo          // 模块列表, 模块会按照注册的顺序序列化
	moduleMap  map[string]*moduleInfo // 模块映射表
}

func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		moduleMap: make(map[string]*moduleInfo),
	}
}

func (mr *ModuleRegistry) Len() int {
	return len(mr.moduleList)
}

// Register 注册模块
func (mr *ModuleRegistry) Register(m Module) *ModuleRegistry {
	if _, ok := mr.moduleMap[m.Key()]; ok {
		panic("module " + m.Key() + " already registered")
	}
	mt := reflect.TypeOf(m)
	if mt.Kind() != reflect.Ptr {
		panic("module " + m.Key() + " must be a pointer")
	}
	mt = mt.Elem()
	mi := &moduleInfo{
		key: m.Key(),
		typ: mt,
	}
	mr.moduleList = append(mr.moduleList, mi)
	mr.moduleMap[mi.key] = mi
	return mr
}

// Create 创建模块实例
func (mr *ModuleRegistry) Create(key string) Module {
	if mi := mr.moduleMap[key]; mi != nil {
		return mi.create()
	}
	return nil
}

// RegisterModule 注册模块的泛型封装.
func RegisterModule[M Module](mr *ModuleRegistry) {
	var m M
	mr.Register(m)
}
