package io

import (
	"context"

	"github.com/godyy/ggs/internal/core/db/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type serverIO struct{}

var Server = &serverIO{}

func init() {
	registerMongoIO(models.MgoCollServer, Server)
}

// createIndexes 创建索引.
func (*serverIO) createIndexes(ctx context.Context, cli *mongo.Client) error {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollServer)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}

// CreateServer 创建服务器
func (*serverIO) CreateServer(ctx context.Context, cli *mongo.Client, server *models.Server) error {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollServer)
	if _, err := coll.InsertOne(ctx, server); err != nil {
		return err
	}
	return nil
}

// GetServer 根据ID获取服务器
func (*serverIO) GetServer(ctx context.Context, cli *mongo.Client, id int64) (*models.Server, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollServer)

	var server models.Server
	if err := coll.FindOne(ctx, bson.M{"id": id}).Decode(&server); err != nil {
		return nil, err
	}

	return &server, nil
}

// GetAllServers 获取所有服务器
func (*serverIO) GetAllServers(ctx context.Context, cli *mongo.Client) ([]*models.Server, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollServer)

	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 0}))
	if err != nil {
		return nil, err
	}

	var servers []*models.Server
	cursor.SetBatchSize(100)
	if err := cursor.All(ctx, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}
