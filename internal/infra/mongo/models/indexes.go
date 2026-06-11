package models

import (
	"context"
	"errors"
	"fmt"

	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var indexesByCol map[string][]mongo.IndexModel

// registerIndexes 注册索引.
func registerIndexes(col string, indexes ...mongo.IndexModel) {
	if indexesByCol == nil {
		indexesByCol = make(map[string][]mongo.IndexModel)
	}
	if _, ok := indexesByCol[col]; ok {
		panic(fmt.Sprintf("indexesByCol[%s] already registered", col))
	}
	if len(indexes) == 0 {
		panic(fmt.Sprintf("indexesByCol[%s] is empty", col))
	}
	indexesByCol[col] = indexes
}

// GetIndexes 获取索引.
func GetIndexes(col string) []mongo.IndexModel {
	if indexesByCol == nil {
		return nil
	}
	return indexesByCol[col]
}

// EnsureIndexes 确保索引存在.
func EnsureIndexes(ctx context.Context, cli *mongo.Client, db string, cols ...string) error {
	if ctx == nil {
		return errors.New("ctx is nil")
	}
	if cli == nil {
		return errors.New("cli is nil")
	}
	if db == "" {
		return errors.New("db is empty")
	}
	indexes := make([][]mongo.IndexModel, 0, len(cols))
	for _, col := range cols {
		colIndexes := GetIndexes(col)
		if len(colIndexes) == 0 {
			return fmt.Errorf("[%s] indexes is empty", col)
		}
		indexes = append(indexes, colIndexes)
	}
	database := cli.Database(db)
	for i := 0; i < len(indexes); i++ {
		collection := database.Collection(cols[i])
		if _, err := collection.Indexes().CreateMany(ctx, indexes[i]); err != nil {
			return pkgerrors.WithMessagef(err, "[%s] create indexes failed", cols[i])
		}
	}
	return nil
}
