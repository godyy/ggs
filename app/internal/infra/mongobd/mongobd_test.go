package mongobd

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	mongocli "github.com/godyy/ggs/internal/base/db/mongo/cli"
	"github.com/godyy/ggs/internal/base/logger"
	"github.com/godyy/ggs/internal/infra/mongobd"
	"github.com/godyy/glog"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestMongoBD(t *testing.T) {
	if err := mongocli.Init(&mongocli.Config{
		URI: "mongodb://localhost:27017/?readPreference=primary",
	}); err != nil {
		t.Fatalf("init mongo failed: %v", err)
	}
	t.Log("init mongo success")

	logger.Init(&logger.Config{
		Level:       glog.DebugLevel,
		Caller:      true,
		Development: true,
		EnableStd:   true,
	})

	db := "test_mongo_bd"
	coll := "test"
	if err := mongocli.Get().Database(db).Drop(context.Background()); err != nil {
		t.Fatalf("drop db failed: %v", err)
	}
	if _, err := mongocli.Get().Database(db).Collection(coll).Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.M{"id": 1},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		t.Fatalf("create index failed: %v", err)
	}

	bd, err := New(Config{
		BDConfig: mongobd.BDConfig{
			Client:       mongocli.Get(),
			Wokers:       runtime.NumCPU(),
			MaxWorkerOps: 10000,
			Logger:       logger.Get(),
		},
		OpChanSize:  10000,
		OpConsumers: 1,
	})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}

	n := 100000
	for i := 0; i < n; i++ {
		op := mongobd.NewOp[mongobd.OpUpdate](db, coll).
			SetFilter(bson.M{"id": i}).
			SetUpdate(bson.M{"id": i, "name": fmt.Sprintf("number_%d", i)}).
			SetUpsert(true)
		if err := bd.Add(i, op, nil); err != nil {
			t.Fatalf("add op failed: %v", err)
		}
	}

	bd.Stop()
}
