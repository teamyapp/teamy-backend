package repo_test

import (
	"fmt"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

type FakeIDRange struct {
	idRangeEnds map[string]int
}

func (f FakeIDRange) GetAllocationEnd(resourceType string) (oneEntity.ID, error) {
	end, ok := f.idRangeEnds[resourceType]
	if !ok {
		return oneEntity.ID(-1), fmt.Errorf("resource type not found")
	}
	return oneEntity.ID(end), nil
}

func (f FakeIDRange) SetAllocationEnd(resourceType string, allocationEnd oneEntity.ID) error {
	f.idRangeEnds[resourceType] = int(allocationEnd)
	return nil
}

var _ repo.IDRange = (*FakeIDRange)(nil)

func NewFakeIDRange() FakeIDRange {
	return FakeIDRange{
		idRangeEnds: make(map[string]int),
	}
}
