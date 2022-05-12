package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type AllocatedRange interface {
	FindByKey(key string) (entity.AllocatedRange, error)
	Add(allocatedRange entity.AllocatedRange) error
	Update(allocatedRange entity.AllocatedRange) error
}
