package repo

import "github.com/godyy/ggskit/base/db/mongo"

func Init(cli *mongo.Client) {
	initMongo(cli)
}
