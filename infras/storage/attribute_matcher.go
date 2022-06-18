package storage

import (
	"errors"
)

type AttributeMatcher struct {
	AttributeName string
	Matcher
}

func (m AttributeMatcher) validate() error {
	if len(m.AttributeName) == 0 {
		return errors.New("attributeName cannot be empty")
	}

	return m.Matcher.Validate()
}

func (m AttributeMatcher) match(attributes map[string]*string) (bool, error) {
	value, ok := attributes[m.AttributeName]
	if !ok {
		return false, nil
	}

	return m.Matcher.match(value)
}
