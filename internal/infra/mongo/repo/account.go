package repo

import (
	"context"

	"github.com/godyy/ggs/internal/infra/mongo/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Account struct {
	col *mongo.Collection
}

func NewAccount(db *mongo.Database) *Account {
	return &Account{
		col: db.Collection(models.CollAccount),
	}
}

// CreateAccount 创建账号.
func (r *Account) CreateAccount(ctx context.Context, account *models.Account) error {
	_, err := r.col.InsertOne(ctx, account)
	return err
}

// GetAccountByUID 根据账号ID获取账号.
func (r *Account) GetAccountByUID(ctx context.Context, uid string) (*models.Account, error) {
	var account models.Account
	err := r.col.FindOne(ctx, bson.M{"uid": uid}).Decode(&account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// CreateOrGetAccount 创建或获取账号：通过uid查找，若不存在则插入，存在则直接返回
func (r *Account) CreateOrGetAccount(ctx context.Context, account *models.Account) (*models.Account, error) {
	var result models.Account
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	filter := bson.M{"uid": account.UID}
	update := bson.M{"$setOnInsert": account} // 仅当插入时写入数据，已存在时不更新
	err := r.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
