package repo

import (
	"github.com/teamyapp/one/entity"
)

type IDRange interface {
	GetAllocationEnd(resourceType string) (entity.ID, error)
	SetAllocationEnd(resourceType string, allocationEnd entity.ID) error
	ListAllResourceTypes() ([]string, error)
}
