package migrate

import (
	"context"

	"github.com/godyy/ggs/internal/base/db/mongo/models"
	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EnsureIndexes 创建索引.
func EnsureIndexes(ctx context.Context, cli *mongo.Client) error {
	for dbname, colIndexes := range indexMap {
		db := cli.Database(dbname)
		for colName, indexes := range colIndexes {
			_, err := db.Collection(colName).Indexes().CreateMany(ctx, indexes)
			if err != nil {
				return pkgerrors.WithMessagef(err, "%s.%s create indexes failed", dbname, colName)
			}
		}
	}
	return nil
}

var (
	indexMap map[string]map[string][]mongo.IndexModel
)

// registerIndexes 注册索引.
func registerIndexes(db, col string, indexes []mongo.IndexModel) {
	if indexMap == nil {
		indexMap = make(map[string]map[string][]mongo.IndexModel)
	}
	if _, ok := indexMap[db]; !ok {
		indexMap[db] = make(map[string][]mongo.IndexModel)
	}
	indexMap[db][col] = indexes
}

func init() {
	registerPlatformIndexes()
}

func registerPlatformIndexes() {
	db := models.MgoDBPlaform
	registerIndexes(db, models.MgoCollAccount,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "id", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys:    bson.D{{Key: "uid", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	registerIndexes(db, models.MgoCollIDGenerator,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "name", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	registerIndexes(db, models.MgoCollCharacter,
		[]mongo.IndexModel{
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
		},
	)
	registerIndexes(db, models.MgoCollServer,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "id", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
}
