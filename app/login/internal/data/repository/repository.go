package repository

import (
	applifecycle "github.com/godyy/ggs/app/internal/base/lifecycle"
	models "github.com/godyy/ggs/internal/base/models/db"
	"github.com/godyy/ggs/internal/data/repository"
	libmongo "github.com/godyy/ggs/internal/libs/db/mongo"
)

var (
	Account     *repository.AccountRepository
	Character   *repository.CharacterRepository
	IDGenerator *repository.IDGeneratorRepository
	Server      *repository.ServerRepository
)

func init() {
	applifecycle.RegisterAfterInitDatabase(func() {
		database := libmongo.Inst().Database(models.MgoDBPlaform)
		Account = repository.NewAccountRepository(database)
		Character = repository.NewCharacterRepository(database)
		IDGenerator = repository.NewIDGeneratorRepository(database)
		Server = repository.NewServerRepository(database)
	})
}
