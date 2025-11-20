package mongobd

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	mongobd3 "github.com/godyy/ggs/internal/core/db/mongobd"

	libmongo "github.com/godyy/ggs/internal/libs/db/mongo"
	liblogger "github.com/godyy/ggs/internal/libs/logger"
	"github.com/godyy/glog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMongoBD(t *testing.T) {
	if err := libmongo.Init(&libmongo.Config{
		URI: "mongodb://localhost:27017/?readPreference=primary",
	}); err != nil {
		t.Fatalf("init mongo failed: %v", err)
	}
	t.Log("init mongo success")

	liblogger.Init(&liblogger.Config{
		Level:       glog.DebugLevel,
		Caller:      true,
		Development: true,
		EnableStd:   true,
	})

	db := "test_mongo_bd"
	coll := "test"
	if err := libmongo.Inst().Database(db).Drop(context.Background()); err != nil {
		t.Fatalf("drop db failed: %v", err)
	}
	if _, err := libmongo.Inst().Database(db).Collection(coll).Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"id": 1},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		t.Fatalf("create index failed: %v", err)
	}

	if err := Start(Config{
		BDConfig: mongobd3.BDConfig{
			Client:       libmongo.Inst(),
			Wokers:       runtime.NumCPU(),
			MaxWorkerOps: 10000,
			Logger:       liblogger.GetLogger(),
		},
		OpChanSize:  10000,
		OpConsumers: 1,
	}); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	n := 100000
	for i := 0; i < n; i++ {
		op := mongobd3.NewOp[mongobd3.OpUpdate](db, coll).
			SetFilter(bson.M{"id": i}).
			SetUpdate(bson.M{"id": i, "name": fmt.Sprintf("number_%d", i)}).
			SetUpsert(true)
		if err := bd.Add(i, op, nil, done); err != nil {
			t.Fatalf("add op failed: %v", err)
		}
	}

	Stop()
}
