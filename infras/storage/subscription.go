package storage

import (
	"errors"
	"fmt"
	"log"
)

type Subscription struct {
	CollectionType    string
	MutationType      MutationType
	AttributeMatchers []AttributeMatcher // must match all the matchers
	ClientID          uint64
}

func (s Subscription) Validate() error {
	if len(s.CollectionType) == 0 {
		return errors.New("collectionType cannot be empty")
	}

	_, ok := mutationTypes[s.MutationType]
	if !ok {
		return fmt.Errorf("invalid mutationType: %v", s.MutationType)
	}

	if len(s.AttributeMatchers) > 0 {
		for _, attributeMatcher := range s.AttributeMatchers {
			err := attributeMatcher.validate()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s Subscription) match(attributes map[string]*string) (bool, error) {
	for _, matcher := range s.AttributeMatchers {
		match, err := matcher.match(attributes)
		if err != nil {
			log.Println(err)
			return false, err
		}

		if !match {
			return false, nil
		}
	}

	return true, nil
}
