package repository

import (
	"context"

	models "github.com/godyy/ggs/internal/base/models/db"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ServerRepository struct {
	col *mongo.Collection
}

func NewServerRepository(db *mongo.Database) *ServerRepository {
	return &ServerRepository{
		col: db.Collection(models.MgoCollServer),
	}
}

// createIndexes 创建索引.
func (s *ServerRepository) createIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	_, err := s.col.Indexes().CreateMany(ctx, indexes)
	return err
}

// CreateServer 创建服务器
func (s *ServerRepository) CreateServer(ctx context.Context, server *models.Server) error {
	if _, err := s.col.InsertOne(ctx, server); err != nil {
		return err
	}
	return nil
}

// GetServer 根据ID获取服务器
func (s *ServerRepository) GetServer(ctx context.Context, id int64) (*models.Server, error) {
	var server models.Server
	if err := s.col.FindOne(ctx, bson.M{"id": id}).Decode(&server); err != nil {
		return nil, err
	}

	return &server, nil
}

// GetAllServers 获取所有服务器
func (s *ServerRepository) GetAllServers(ctx context.Context) ([]*models.Server, error) {
	cursor, err := s.col.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 0}))
	if err != nil {
		return nil, err
	}

	var servers []*models.Server
	cursor.SetBatchSize(100)
	if err := cursor.All(ctx, &servers); err != nil {
		return nil, err
	}

	return servers, nil
}
