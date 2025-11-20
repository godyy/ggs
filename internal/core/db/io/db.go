package io

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// mongoIO 定义MongoDB io 模块.
type mongoIO interface {
	// createIndexes 创建索引.
	createIndexes(ctx context.Context, cli *mongo.Client) error
}

// mongoIOs 维护mongoIO模块
var mongoIOs = map[string]mongoIO{}

// registerMongoIO 注册mongoIO模块.
func registerMongoIO(name string, io mongoIO) {
	if _, ok := mongoIOs[name]; ok {
		panic("mongoIO " + name + " already registered")
	}
	mongoIOs[name] = io
}

// InitMongoIndexes 初始化MongoDB索引.
func InitMongoIndexes(ctx context.Context, cli *mongo.Client) error {
	for _, io := range mongoIOs {
		if err := io.createIndexes(ctx, cli); err != nil {
			return err
		}
	}
	return nil
}
