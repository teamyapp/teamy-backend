package dao

import "github.com/teamyapp/cloud/app/entity"

type OperationRelation interface {
	FindOperationRelations(childResourceType string, childOperation string) ([]entity.OperationRelation, error)
}
