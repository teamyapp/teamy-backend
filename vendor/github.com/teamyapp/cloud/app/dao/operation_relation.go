package dao

import "github.com/teamyapp/cloud/app/entity"

type OperationRelation interface {
	FindOperationRelation(
		childResourceType string,
		childOperation string,
		parentResourceType string,
		parentOperation string,
	) (entity.OperationRelation, error)
	FindOperationRelations(childResourceType string, childOperation string) ([]entity.OperationRelation, error)
	FindAllOperationRelations() ([]entity.OperationRelation, error)
	CreateOperationRelation(operationRelation entity.OperationRelation) error
	DeleteOperationRelation(
		childResourceType string,
		childOperation string,
		parentResourceType string,
		parentOperation string,
	) error
}
