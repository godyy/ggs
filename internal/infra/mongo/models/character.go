package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Character 角色.
type Character struct {
	ID        int64  `bson:"id"`         // 角色ID
	AccountID int64  `bson:"account_id"` // 账号ID
	Name      string `bson:"name"`       // 角色名称
	ServerID  int64  `bson:"server_id"`  // 服务器ID
}

func init() {
	registerIndexes(CollCharacter,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		mongo.IndexModel{
			Keys:    bson.D{{Key: "account_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
		mongo.IndexModel{
			Keys:    bson.D{{Key: "account_id", Value: 1}, {Key: "server_id", Value: 1}},
			Options: options.Index().SetUnique(false),
		},
	)
}
