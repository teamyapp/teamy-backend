package idgen

import (
	"fmt"
	"math"

	"github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

const maxUserID = math.MaxInt32

type idRange struct {
	rangeStart   entity.ID
	rangeEnd     entity.ID
	nextUniqueID entity.ID
}

type IDGenerator struct {
	allocatedRanges map[string]*idRange
	idRangeRepo     repo.IDRange
	rangeLength     int
}

func (i IDGenerator) NextUniqueID(resourceType string) (entity.ID, error) {
	if _, ok := i.allocatedRanges[resourceType]; !ok {
		return entity.ID(-1), fmt.Errorf("unregistered resourceType: %s", resourceType)
	}
	alloc := i.allocatedRanges[resourceType]
	if alloc.nextUniqueID > alloc.rangeEnd {
		newRange, err := i.allocateIDRange(resourceType)
		if err != nil {
			return entity.ID(-1), err
		}
		alloc = &newRange
	}
	id := alloc.nextUniqueID
	alloc.nextUniqueID++
	i.allocatedRanges[resourceType] = alloc
	return id, nil
}

func (i IDGenerator) RegisterResourceType(resourceType string) error {
	err := i.idRangeRepo.SetAllocationEnd(resourceType, -1)
	if err != nil {
		return err
	}

	newRange, err := i.allocateIDRange(resourceType)
	if err != nil {
		return err
	}
	i.allocatedRanges[resourceType] = &newRange
	return nil
}

func (i IDGenerator) UnregisterResourceType(resourceType string) {
	delete(i.allocatedRanges, resourceType)
}

func (i IDGenerator) allocateIDRange(resourceType string) (idRange, error) {
	// TODO: partition based on resource type for distributed systems
	rangeEnd, err := i.idRangeRepo.GetAllocationEnd(resourceType)
	if err != nil {
		return idRange{}, err
	}
	if rangeEnd == maxUserID {
		return idRange{}, fmt.Errorf("out of ID to allocate")
	}
	newRangeStart := rangeEnd + 1
	newRangeEnd := max(int(rangeEnd) + i.rangeLength - 1, maxUserID)
	err = i.idRangeRepo.SetAllocationEnd(resourceType, rangeEnd)
	return idRange{
		rangeStart:   newRangeStart,
		rangeEnd:     entity.ID(newRangeEnd),
		nextUniqueID: newRangeStart,
	}, err
}

func max(num1 int, num2 int) int {
	if num1 >= num2 {
		return num1
	} else {
		return num2
	}
}

func newIDGenerator(idRangeRepo repo.IDRange, rangeLength int) IDGenerator {
	return IDGenerator{
		allocatedRanges: make(map[string]*idRange),
		idRangeRepo:     idRangeRepo,
		rangeLength:     rangeLength,
	}
}
