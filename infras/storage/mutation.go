package storage

import (
	"errors"
	"fmt"
)

type MutationType string

const (
	CreateMutationType MutationType = "CREATE"
	UpdateMutationType MutationType = "UPDATE"
	DeleteMutationType MutationType = "DELETE"
)

var mutationTypes = map[MutationType]bool{
	CreateMutationType: true,
	UpdateMutationType: true,
	DeleteMutationType: true,
}

type Mutation struct {
	CollectionType string
	MutationType   MutationType
	Attributes     map[string]*string
}

func (m Mutation) Validate() error {
	if len(m.CollectionType) == 0 {
		return errors.New("collectionType cannot be empty")
	}

	_, ok := mutationTypes[m.MutationType]
	if !ok {
		return fmt.Errorf("invalid mutationType: %v", m.MutationType)
	}

	return nil
}

func (m Mutation) String() string {
	return fmt.Sprintf("Mutation[CollectionType:%v MutationType:%v Attributes:%v]",
		m.CollectionType, m.MutationType, replacePointers(m.Attributes))
}

func replacePointers(input map[string]*string) map[string]string {
	output := make(map[string]string)
	for key, value := range input {
		if value == nil {
			output[key] = "<nil>"
		} else {
			output[key] = *value
		}
	}

	return output
}
