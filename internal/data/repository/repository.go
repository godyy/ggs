package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// dao 定义MongoDB dao 模块.
type dao interface {
	// createIndexes 创建索引.
	createIndexes(ctx context.Context, cli *mongo.Client) error
}

// daos 维护mongoIO模块
var daos = map[string]dao{}

// registerDAO 注册dao模块.
func registerDAO(name string, io dao) {
	if _, ok := daos[name]; ok {
		panic("dao " + name + " already registered")
	}
	daos[name] = io
}

// CreateIndexes 初始化MongoDB索引.
func CreateIndexes(ctx context.Context, cli *mongo.Client) error {
	for _, io := range daos {
		if err := io.createIndexes(ctx, cli); err != nil {
			return err
		}
	}
	return nil
}
