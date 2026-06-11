package repo

import (
	"context"

	"github.com/godyy/ggs/internal/infra/mongo/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IDGenerator struct {
	col *mongo.Collection
}

func NewIDGenerator(db *mongo.Database) *IDGenerator {
	return &IDGenerator{
		col: db.Collection(models.CollIDGenerator),
	}
}

// GenID 生成ID.
func (r *IDGenerator) GenID(ctx context.Context, name string) (int64, error) {
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
func (r *IDGenerator) GenAccountID(ctx context.Context) (int64, error) {
	return r.GenID(ctx, models.IDGenAccount)
}

// GenCharacterID 生成角色ID.
func (r *IDGenerator) GenCharacterID(ctx context.Context) (int64, error) {
	return r.GenID(ctx, models.IDGenCharacter)
}
