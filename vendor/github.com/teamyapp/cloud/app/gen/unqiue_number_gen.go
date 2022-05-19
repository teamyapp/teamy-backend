package gen

import (
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type UniqueNumber struct {
	allocatedRangeDao dao.AllocatedRange
	name              string
	rangeSize         uint64
	allocatedRange    entity.AllocatedRange
}

func (u *UniqueNumber) GenerateUniqueNumber() (uint64, error) {
	if u.allocatedRange.NextNumber > u.allocatedRange.RangeEnd {
		err := u.allocateNewRange()
		if err != nil {
			return uint64(0), err
		}
	}

	num := u.allocatedRange.NextNumber
	u.allocatedRange.NextNumber++
	return num, nil
}

func (u *UniqueNumber) allocateNewRange() error {
	if u.allocatedRange.RangeEnd == math.MaxInt64 {
		return fmt.Errorf("out of number to allocate")
	}

	newRangeStart := u.allocatedRange.RangeEnd + 1
	newRangeEnd := min(u.allocatedRange.RangeEnd+u.rangeSize, math.MaxUint64)
	newRange := entity.AllocatedRange{
		Key:        u.name,
		RangeEnd:   newRangeEnd,
		NextNumber: newRangeStart,
	}
	err := u.allocatedRangeDao.Update(newRange)
	if err != nil {
		return err
	}

	u.allocatedRange = newRange
	log.Printf("allocated range: %v", newRange)
	return nil
}

func min[Number int | uint64](num1 Number, num2 Number) Number {
	if num1 <= num2 {
		return num1
	} else {
		return num2
	}
}

func newUniqueNumber(
	allocatedRangeDao dao.AllocatedRange,
	name string,
	rangeSize uint64,
) (*UniqueNumber, error) {
	allocatedRange, err := allocatedRangeDao.FindByKey(name)
	var errNotFound dao.ErrNotFound
	if err != nil {
		if !errors.As(err, &errNotFound) {
			return nil, err
		}

		allocatedRange = entity.AllocatedRange{
			Key:        name,
			RangeEnd:   0,
			NextNumber: 0,
		}

		err = allocatedRangeDao.Add(allocatedRange)
		if err != nil {
			return nil, err
		}
	}

	uniqueNumber := &UniqueNumber{
		name:              name,
		rangeSize:         rangeSize,
		allocatedRange:    allocatedRange,
		allocatedRangeDao: allocatedRangeDao,
	}
	err = uniqueNumber.allocateNewRange()
	return uniqueNumber, err
}

type UniqueNumberFactory struct {
	allocatedRangeDao dao.AllocatedRange
	rangeSize         uint64
}

func (u UniqueNumberFactory) MakeUniqueNumber(name string) (*UniqueNumber, error) {
	return newUniqueNumber(u.allocatedRangeDao, name, u.rangeSize)
}

func NewUniqueNumberFactory(allocatedRangeDao dao.AllocatedRange, rangeSize uint64) UniqueNumberFactory {
	return UniqueNumberFactory{
		allocatedRangeDao: allocatedRangeDao,
		rangeSize:         rangeSize,
	}
}
