package collection

import (
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/infras/storage"
)

const userCollectionType = "User"

type UserSyncer struct {
	realTimeCollection *storage.RealTimeCollections
	userDao            dao.User
}

func (u UserSyncer) CreateAndSyncUser(user entity.User) error {
	err := u.userDao.CreateUser(user)
	if err != nil {
		return err
	}

	return u.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: userCollectionType,
		MutationType:   storage.CreateMutationType,
		Attributes:     storage.MapAttributes(user),
	})
}

func (u UserSyncer) UpdateAndSyncUser(user entity.User) error {
	err := u.userDao.UpdateUser(user)
	if err != nil {
		return err
	}

	return u.realTimeCollection.Mutate(storage.Mutation{
		CollectionType: userCollectionType,
		MutationType:   storage.UpdateMutationType,
		Attributes:     storage.MapAttributes(user),
	})
}

func NewUserSyncer(realTimeCollection *storage.RealTimeCollections, userDao dao.User) UserSyncer {
	realTimeCollection.RegisterCollectionType(userCollectionType)
	return UserSyncer{
		realTimeCollection: realTimeCollection,
		userDao:            userDao,
	}
}
