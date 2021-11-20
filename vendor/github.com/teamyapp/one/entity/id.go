package entity

import (
	"fmt"
)

type ID int

func (i ID) IsValid() error {
	if i <= 0 {
		return fmt.Errorf("ID must be positive: %d", i)
	}

	return nil
}
