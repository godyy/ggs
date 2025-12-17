package repo

import (
	mongocli "github.com/godyy/ggs/internal/base/db/mongo/cli"
	mongomodels "github.com/godyy/ggs/internal/base/db/mongo/models"
	mongorepo "github.com/godyy/ggs/internal/base/db/mongo/repo"
)

var (
	Account     *mongorepo.Account
	Character   *mongorepo.Character
	IDGenerator *mongorepo.IDGenerator
	Server      *mongorepo.Server
)

func initMongo() {
	cli := mongocli.Get()
	platDB := cli.Database(mongomodels.MgoDBPlaform)
	Account = mongorepo.NewAccount(platDB)
	Character = mongorepo.NewCharacter(platDB)
	IDGenerator = mongorepo.NewIDGenerator(platDB)
	Server = mongorepo.NewServer(platDB)
}
