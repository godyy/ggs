package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Account 账号
type Account struct {
	ID  int64  `bson:"id"`  // 账号ID
	UID string `bson:"uid"` // 用户ID
}

func init() {
	registerIndexes(CollAccount,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		mongo.IndexModel{
			Keys:    bson.D{{Key: "uid", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
}
