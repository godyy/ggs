package repo

import (
	mongomodels "github.com/godyy/ggs/internal/infra/mongo/models"
	mongorepo "github.com/godyy/ggs/internal/infra/mongo/repo"
	"github.com/godyy/ggskit/base/db/mongo"
)

var (
	Account     *mongorepo.Account
	Character   *mongorepo.Character
	IDGenerator *mongorepo.IDGenerator
	Server      *mongorepo.Server
)

func initMongo(cli *mongo.Client) {
	platDB := cli.Database(mongomodels.DBPlatform)
	Account = mongorepo.NewAccount(platDB)
	Character = mongorepo.NewCharacter(platDB)
	IDGenerator = mongorepo.NewIDGenerator(platDB)
	Server = mongorepo.NewServer(platDB)
}
