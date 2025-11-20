package io

import (
	"context"

	"github.com/godyy/ggs/internal/core/db/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type accountIO struct{}

var Account = &accountIO{}

func init() {
	registerMongoIO(models.MgoCollAccount, Account)
}

// createIndexes 创建索引.
func (*accountIO) createIndexes(ctx context.Context, cli *mongo.Client) error {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollAccount)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "uid", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}

// CreateAccount 创建账号.
func (*accountIO) CreateAccount(ctx context.Context, cli *mongo.Client, account *models.Account) error {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollAccount)
	_, err := coll.InsertOne(ctx, account)
	return err
}

// GetAccountByUID 根据账号ID获取账号.
func (*accountIO) GetAccountByUID(ctx context.Context, cli *mongo.Client, uid string) (*models.Account, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollAccount)
	var account models.Account
	err := coll.FindOne(ctx, bson.M{"uid": uid}).Decode(&account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// CreateOrGetAccount 创建或获取账号：通过uid查找，若不存在则插入，存在则直接返回
func (*accountIO) CreateOrGetAccount(ctx context.Context, cli *mongo.Client, account *models.Account) (*models.Account, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollAccount)
	var result models.Account
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{"uid": account.UID}
	update := bson.M{"$setOnInsert": account} // 仅当插入时写入数据，已存在时不更新
	err := coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
