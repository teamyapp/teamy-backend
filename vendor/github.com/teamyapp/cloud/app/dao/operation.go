package dao

import "github.com/teamyapp/cloud/app/entity"

type Operation interface {
	FindOperation(resourceTypeName string, operationName string) (entity.Operation, error)
	FindAllOperations() ([]entity.Operation, error)
	CreateOperation(operation entity.Operation) error
	DeleteOperation(resourceTypeName string, operationName string) error
}
