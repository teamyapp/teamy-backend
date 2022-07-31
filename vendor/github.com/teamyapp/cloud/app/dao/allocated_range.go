package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type AllocatedRange interface {
	FindAllocatedRangeByKey(key string) (entity.AllocatedRange, error)
	CreateAllocatedRange(allocatedRange entity.AllocatedRange) error
	UpdateAllocatedRange(allocatedRange entity.AllocatedRange) error
}
