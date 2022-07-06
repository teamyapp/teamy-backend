package dao

import (
	"github.com/teamyapp/cloud/app/entity"
)

type ServiceAccount interface {
	FindAllServiceAccounts(accountOwnerID uint64) ([]entity.ServiceAccount, error)
	FindServiceAccountByID(serviceAccountID uint64) (entity.ServiceAccount, error)
	CreateServiceAccount(serviceAccount entity.ServiceAccount) error
	UpdateServiceAccount(serviceAccount entity.ServiceAccount) error
	DeleteServiceAccount(serviceAccountID uint64) error
}
