package entity

import (
	"fmt"
)

type AllocatedRange struct {
	Key        string
	RangeEnd   uint64
	NextNumber uint64
}

func (a AllocatedRange) String() string {
	return fmt.Sprintf("[AllocatedRange Key=%v RangeEnd=%v NextNumber=%v]", a.Key, a.RangeEnd, a.NextNumber)
}
