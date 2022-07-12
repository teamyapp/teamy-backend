package dao

import "github.com/teamyapp/cloud/app/entity"

type Permission interface {
	FindPermission(query entity.PermissionQuery) (entity.Permission, error)
}
