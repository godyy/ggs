package model

import (
	"context"
	"encoding/hex"
	"reflect"
	"testing"

	libmongo "github.com/godyy/ggs/internal/libs/db/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type testActor struct{}

func (ta *testActor) OnModelDirty() {
}

type testModuleBase[M Module] = ModuleBase[M]

type testModel struct {
	mr          *ModuleRegistry
	ID          int64      `bson:"id"`
	Name        string     `bson:"name"`
	Modules     *ModuleMgr `bson:"modules"`
	*DirtyModel `bson:"-"`
}

func newTestModel(mr *ModuleRegistry) *testModel {
	m := &testModel{
		mr:         mr,
		DirtyModel: NewDirtyModel(&testActor{}),
	}
	m.Modules = NewModuleMgr(m)
	return m
}

func (m *testModel) ModuleRegistry() *ModuleRegistry {
	return m.mr
}

func (m *testModel) SetModuleDirty(mm Module) {
	m.DirtyModel.SetDirty("modules."+mm.Key(), mm)
}

func (m *testModel) GetModule(key string, autoCreate bool) Module {
	return m.Modules.GetModule(key, autoCreate)
}

func (m *testModel) GetHashKey() any { return m.ID }

func (m *testModel) GetCollection() string { return "test_models" }

func (m *testModel) GetFilter() any {
	return bson.M{"id": m.ID}
}

func (m *testModel) Release() {
	m.DirtyModel.ClearDirty()
	m.Modules.Release()
	m.mr = nil
}

type testModuleA struct {
	testModuleBase[*testModuleA]
	Value string
}

func (m *testModuleA) Key() string { return "A" }

type testModuleB struct {
	testModuleBase[*testModuleB]
	Value string
}

func (m *testModuleB) Key() string { return "B" }

type testModuleKeySA struct{}

func (k testModuleKeySA) Key() string { return "SA" }

type testSA struct {
	Value string
}

type testModuleSA = ModuleSingle[*string, testModuleKeySA]

func testPValue[V any](v V) *V {
	return &v
}

func TestModulesCodec(t *testing.T) {
	registry := NewModuleRegistry()
	RegisterModule[*testModuleB](registry)
	RegisterModule[*testModuleA](registry)
	RegisterModule[*testModuleSA](registry)

	modelSrc := newTestModel(registry)
	modelSrc.ID = 1
	modelSrc.Name = "test"
	modelSrc.Modules.InitAllModules()
	GetModule[*testModuleA](modelSrc, false).Value = "this is module A"
	GetModule[*testModuleB](modelSrc, false).Value = "this is module B"
	GetModule[*testModuleSA](modelSrc, false).Set(testPValue("123"))
	modelSrcBSON, err := bson.Marshal(modelSrc)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(modelSrcBSON))

	modelDst := newTestModel(registry)
	if err := bson.Unmarshal(modelSrcBSON, modelDst); err != nil {
		t.Fatal(err)
	}

	modelDstBSON, err := bson.Marshal(modelDst)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(modelDstBSON))

	if !reflect.DeepEqual(modelDstBSON, modelSrcBSON) {
		t.Fatalf("dst:%+v not equal src:%+v", modelDst, modelSrc)
	}

	GetModule[*testModuleA](modelDst, false).Value = "this is module AA"
	GetModule[*testModuleA](modelDst, false).SetDirty()
	GetModule[*testModuleB](modelDst, false).Value = "this is module BB"
	GetModule[*testModuleB](modelDst, false).SetDirty()
	if dirty, _ := modelDst.DirtyModel.IsDirty(); !dirty {
		t.Fatal("actorDst.DirtyMgr not dirty")
	}

	GetModule[*testModuleA](modelDst, false).Value = "this is module AAA"
	GetModule[*testModuleA](modelDst, false).SetDirty()
	GetModule[*testModuleB](modelDst, false).Value = "this is module BBB"
	GetModule[*testModuleB](modelDst, false).SetDirty()
	modelDstBSON, err = modelDst.MarshalBSONDirty()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(modelDstBSON))
}

func TestModulesDirty(t *testing.T) {
	if err := libmongo.Init(&libmongo.Config{
		URI: "mongodb://localhost:27017/?readPreference=primary",
	}); err != nil {
		t.Fatal(err)
	}
	db := libmongo.Inst().Database("test")
	coll := db.Collection("test_models")

	registry := NewModuleRegistry()
	RegisterModule[*testModuleB](registry)
	RegisterModule[*testModuleA](registry)

	model := newTestModel(registry)
	model.ID = 1
	model.Name = "test"
	model.Modules.InitAllModules()
	if _, err := coll.InsertOne(context.Background(), model); err != nil {
		t.Fatal(err)
	}

	GetModule[*testModuleA](model, false).Value = "this is module A"
	GetModule[*testModuleA](model, false).SetDirty()
	// GetModule[*testModuleB](model).Value = "this is module B"
	// GetModule[*testModuleB](model).SetDirty()
	modelDirtyBSON, err := model.MarshalBSONDirty()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hex.EncodeToString(modelDirtyBSON))

	if _, err := coll.UpdateOne(context.Background(),
		bson.M{"id": model.ID},
		bson.M{"$set": bson.Raw(modelDirtyBSON)},
	); err != nil {
		t.Fatal(err)
	}

}
