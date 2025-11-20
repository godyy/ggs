package io

import (
	"context"

	"github.com/godyy/ggs/internal/core/db/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type idGeneratorIO struct{}

var IDGenerator = &idGeneratorIO{}

func init() {
	registerMongoIO(models.MgoCollIDGenerator, IDGenerator)
}

func (*idGeneratorIO) createIndexes(ctx context.Context, cli *mongo.Client) error {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollIDGenerator)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}

// GenID 生成ID.
func (*idGeneratorIO) GenID(ctx context.Context, cli *mongo.Client, name string) (int64, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollIDGenerator)
	filter := bson.D{{Key: "name", Value: name}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "counter", Value: 1}}}}
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetUpsert(true).
		SetProjection(bson.D{{Key: "counter", Value: 1}})
	var result struct {
		Counter int64 `bson:"counter"`
	}
	if err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result); err != nil {
		return 0, err
	}
	return result.Counter, nil
}

// GenAccountID 生成账号ID.
func (*idGeneratorIO) GenAccountID(ctx context.Context, cli *mongo.Client) (int64, error) {
	return IDGenerator.GenID(ctx, cli, models.MgoIDGeneratorAccount)
}

// GenCharacterID 生成角色ID.
func (*idGeneratorIO) GenCharacterID(ctx context.Context, cli *mongo.Client) (int64, error) {
	return IDGenerator.GenID(ctx, cli, models.MgoIDGeneratorCharacter)
}
