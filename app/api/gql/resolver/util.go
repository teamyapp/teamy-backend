package resolver

import (
	oneEntity "github.com/teamyapp/one/entity"
)

func contains(arr []oneEntity.ID, element oneEntity.ID) bool {
	for _, e := range arr {
		if e == element {
			return true
		}
	}
	return false
}
