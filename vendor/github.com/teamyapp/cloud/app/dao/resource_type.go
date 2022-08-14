package dao

import "github.com/teamyapp/cloud/app/entity"

type ResourceType interface {
	FindResourceType(resourceType string) (entity.ResourceType, error)
	FindAllResourceTypes() ([]entity.ResourceType, error)
	CreateResourceType(resourceTypeEntity entity.ResourceType) error
	DeleteResourceType(resourceType string) error
}
