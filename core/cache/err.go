package cache

import "fmt"

type KeyNotFoundErr[Key comparable] struct {
	Key Key
}

var _ error = (*KeyNotFoundErr[string])(nil)

func (k KeyNotFoundErr[Key]) Error() string {
	return fmt.Sprintf("key not found: %v", k.Key)
}

type InvalidCapacityErr struct {
	Capacity int
}

var _ error = (*InvalidCapacityErr)(nil)

func (i InvalidCapacityErr) Error() string {
	return fmt.Sprintf("invalid capacity: %v", i.Capacity)
}
