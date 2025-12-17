package migrate

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Migrate(ctx context.Context, cli *mongo.Client) error {
	if err := EnsureIndexes(ctx, cli); err != nil {
		return err
	}
	return nil
}
