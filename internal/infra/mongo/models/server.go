package models

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Server 服务器
type Server struct {
	ID     int64  `bson:"id"`     // 服务器ID
	Name   string `bson:"name"`   // 服务器名称
	NodeId string `bson:"nodeId"` // 服务器所在节点ID
}

func init() {
	registerIndexes(CollServer, mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}
