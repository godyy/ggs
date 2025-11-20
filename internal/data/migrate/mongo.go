package migrate

import (
	"context"

	models "github.com/godyy/ggs/internal/base/models/db"
	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Mongo(ctx context.Context, cli *mongo.Client) error {
	return MongoEnsureIndexes(ctx, cli)
}

func MongoEnsureIndexes(ctx context.Context, cli *mongo.Client) error {
	for dbname, colIndexes := range mongoIndexes {
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
	mongoIndexes map[string]map[string][]mongo.IndexModel
)

func registerMongoIndexes(db, col string, indexes []mongo.IndexModel) {
	if mongoIndexes == nil {
		mongoIndexes = make(map[string]map[string][]mongo.IndexModel)
	}
	if _, ok := mongoIndexes[db]; !ok {
		mongoIndexes[db] = make(map[string][]mongo.IndexModel)
	}
	mongoIndexes[db][col] = indexes
}

func init() {
	registerMongoPlatformIndexes()
}

func registerMongoPlatformIndexes() {
	db := models.MgoDBPlaform
	registerMongoIndexes(db, models.MgoCollAccount,
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
	registerMongoIndexes(db, models.MgoCollIDGenerator,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "name", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	registerMongoIndexes(db, models.MgoCollCharacter,
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
	registerMongoIndexes(db, models.MgoCollServer,
		[]mongo.IndexModel{
			{
				Keys:    bson.D{{Key: "id", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
}
