package idgen

import (
	"fmt"
	"github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
	"log"
	"math"
	"sync"
)

const maxID = math.MaxInt32

type idRange struct {
	rangeEnd     entity.ID
	nextUniqueID entity.ID
	mutex sync.Mutex
}

type IDGenerator struct {
	allocatedRanges map[string]*idRange
	idRangeRepo     repo.IDRange
	rangeLength     int
	mutex sync.RWMutex
}

func (i *IDGenerator) NextUniqueID(resourceType string) (entity.ID, error) {
	if _, ok := i.allocatedRanges[resourceType]; !ok {
		return entity.ID(-1), fmt.Errorf("unregistered resourceType: %s", resourceType)
	}

	i.mutex.RLock()
	alloc := i.allocatedRanges[resourceType]
	i.mutex.RUnlock()

	alloc.mutex.Lock()
	defer alloc.mutex.Unlock()
	if alloc.nextUniqueID > alloc.rangeEnd {
		alloc.mutex.Unlock()
		newRange, err := i.allocateIDRange(resourceType)
		if err != nil {
			return entity.ID(-1), err
		}
		alloc.mutex.Lock()
		alloc = &newRange
	}
	id := alloc.nextUniqueID
	alloc.nextUniqueID++
	return id, nil
}

func (i *IDGenerator) RegisterResourceType(resourceType string) error {
	err := i.idRangeRepo.SetAllocationEnd(resourceType, -1)
	if err != nil {
		return err
	}

	newRange, err := i.allocateIDRange(resourceType)
	if err != nil {
		return err
	}

	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.allocatedRanges[resourceType] = &newRange
	return nil
}

func (i *IDGenerator) UnregisterResourceType(resourceType string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	delete(i.allocatedRanges, resourceType)
}

func (i *IDGenerator) Init() error {
	resourceTypes, err := i.idRangeRepo.ListAllResourceTypes()
	if err != nil {
		return err
	}

	for _, resourceType := range resourceTypes {
		ir, err := i.allocateIDRange(resourceType)
		if err != nil {
			log.Printf("cannot read id range from database: resource type=%s", resourceType)
			continue
		}
		i.allocatedRanges[resourceType] = &ir
	}
	return nil
}

func (i *IDGenerator) allocateIDRange(resourceType string) (idRange, error) {
	// TODO: partition based on resource type for distributed systems
	rangeEnd, err := i.idRangeRepo.GetAllocationEnd(resourceType)
	if err != nil {
		return idRange{}, err
	}
	if rangeEnd == maxID {
		return idRange{}, fmt.Errorf("out of ID to allocate")
	}
	newRangeStart := rangeEnd + 1
	newRangeEnd := min(int(rangeEnd) + i.rangeLength - 1, maxID)
	err = i.idRangeRepo.SetAllocationEnd(resourceType, rangeEnd)
	return idRange{
		rangeEnd:     entity.ID(newRangeEnd),
		nextUniqueID: newRangeStart,
	}, err
}

func min(num1 int, num2 int) int {
	if num1 <= num2 {
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
