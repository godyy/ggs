package io

import (
	"context"

	"github.com/godyy/ggs/internal/core/db/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type characterIO struct{}

var Character = &characterIO{}

func init() {
	registerMongoIO(models.MgoCollCharacter, Character)
}

// createIndexes 创建索引.
func (*characterIO) createIndexes(ctx context.Context, cli *mongo.Client) error {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollCharacter)
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "account_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: "account_id", Value: 1}, {Key: "server_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}

// CreateCharacter 创建角色
func (c *characterIO) CreateCharacter(ctx context.Context, cli *mongo.Client, character *models.Character) error {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollCharacter)
	_, err := coll.InsertOne(ctx, character)
	return err
}

// GetCharacter 根据角色ID获取角色
func (c *characterIO) GetCharacter(ctx context.Context, cli *mongo.Client, characterID int64) (*models.Character, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollCharacter)

	var character models.Character
	err := coll.FindOne(ctx, bson.M{"id": characterID}).Decode(&character)
	if err != nil {
		return nil, err
	}

	return &character, nil
}

// GetAllCharactersByAccountID 根据账号ID获取所有角色
func (c *characterIO) GetAllCharactersByAccountID(ctx context.Context, cli *mongo.Client, accountID int64) ([]*models.Character, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollCharacter)

	cursor, err := coll.Find(ctx, bson.M{"account_id": accountID}, options.Find().SetProjection(bson.M{"_id": 0}))
	if err != nil {
		return nil, err
	}

	var characters []*models.Character
	cursor.SetBatchSize(100)
	if err := cursor.All(ctx, &characters); err != nil {
		return nil, err
	}

	return characters, nil
}

// GetCharacterCountByAccountID 根据账号ID获取角色数量
func (c *characterIO) GetCharacterCountByAccountID(ctx context.Context, cli *mongo.Client, accountID int64) (int64, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollCharacter)

	count, err := coll.CountDocuments(ctx, bson.M{"account_id": accountID})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetCharacterCountByAccounIDServerID 根据账号ID和服务器ID获取角色数量
func (c *characterIO) GetCharacterCountByAccounIDServerID(ctx context.Context, cli *mongo.Client, accountID int64, serverID int64) (int64, error) {
	coll := cli.Database(models.MgoDBPlaform).Collection(models.MgoCollCharacter)

	count, err := coll.CountDocuments(ctx, bson.M{"account_id": accountID, "server_id": serverID})
	if err != nil {
		return 0, err
	}

	return count, nil
}
