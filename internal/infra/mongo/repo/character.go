package repo

import (
	"context"

	"github.com/godyy/ggs/internal/infra/mongo/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Character struct {
	col *mongo.Collection
}

func NewCharacter(db *mongo.Database) *Character {
	return &Character{
		col: db.Collection(models.CollCharacter),
	}
}

// CreateCharacter 创建角色
func (c *Character) CreateCharacter(ctx context.Context, character *models.Character) error {
	_, err := c.col.InsertOne(ctx, character)
	return err
}

// GetCharacter 根据角色ID获取角色
func (c *Character) GetCharacter(ctx context.Context, characterID int64) (*models.Character, error) {
	var character models.Character
	err := c.col.FindOne(ctx, bson.M{"id": characterID}).Decode(&character)
	if err != nil {
		return nil, err
	}

	return &character, nil
}

// GetAllCharactersByAccountID 根据账号ID获取所有角色
func (c *Character) GetAllCharactersByAccountID(ctx context.Context, accountID int64) ([]*models.Character, error) {
	cursor, err := c.col.Find(ctx, bson.M{"account_id": accountID}, options.Find().SetProjection(bson.M{"_id": 0}))
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
func (c *Character) GetCharacterCountByAccountID(ctx context.Context, accountID int64) (int64, error) {
	count, err := c.col.CountDocuments(ctx, bson.M{"account_id": accountID})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetCharacterCountByAccounIDServerID 根据账号ID和服务器ID获取角色数量
func (c *Character) GetCharacterCountByAccounIDServerID(ctx context.Context, accountID int64, serverID int64) (int64, error) {
	count, err := c.col.CountDocuments(ctx, bson.M{"account_id": accountID, "server_id": serverID})
	if err != nil {
		return 0, err
	}

	return count, nil
}
