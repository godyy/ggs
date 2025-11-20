package repository

import (
	"context"

	models "github.com/godyy/ggs/internal/base/models/db"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IDGeneratorRepository struct {
	col *mongo.Collection
}

func NewIDGeneratorRepository(db *mongo.Database) *IDGeneratorRepository {
	return &IDGeneratorRepository{
		col: db.Collection(models.MgoCollIDGenerator),
	}
}

func (r *IDGeneratorRepository) createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := r.col.Indexes().CreateMany(ctx, indexes)
	return err
}

// GenID 生成ID.
func (r *IDGeneratorRepository) GenID(ctx context.Context, name string) (int64, error) {
	filter := bson.D{{Key: "name", Value: name}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "counter", Value: 1}}}}
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetUpsert(true).
		SetProjection(bson.D{{Key: "counter", Value: 1}})
	var result struct {
		Counter int64 `bson:"counter"`
	}
	if err := r.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result); err != nil {
		return 0, err
	}
	return result.Counter, nil
}

// GenAccountID 生成账号ID.
func (r *IDGeneratorRepository) GenAccountID(ctx context.Context) (int64, error) {
	return r.GenID(ctx, models.MgoIDGeneratorAccount)
}

// GenCharacterID 生成角色ID.
func (r *IDGeneratorRepository) GenCharacterID(ctx context.Context) (int64, error) {
	return r.GenID(ctx, models.MgoIDGeneratorCharacter)
}
