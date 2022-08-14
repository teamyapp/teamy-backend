package dao

import "github.com/teamyapp/cloud/app/entity"

type ResourceRelation interface {
	FindResourceRelation(
		childResourceType string,
		childResourceID uint64,
		parentResourceType string,
		parentResourceID uint64,
	) (entity.ResourceRelation, error)
	FindResourceRelations(childResourceType string, childResourceID uint64) ([]entity.ResourceRelation, error)
	FindAllResourceRelations() ([]entity.ResourceRelation, error)
	CreateResourceRelation(resourceRelation entity.ResourceRelation) error
	DeleteResourceRelation(
		childResourceType string,
		childResourceID uint64,
		parentResourceType string,
		parentResourceID uint64,
	) error
}
