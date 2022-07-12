package dao

import "github.com/teamyapp/cloud/app/entity"

type ResourceRelation interface {
	FindResourceRelations(childResourceID uint64, childResourceType string) ([]entity.ResourceRelation, error)
}
