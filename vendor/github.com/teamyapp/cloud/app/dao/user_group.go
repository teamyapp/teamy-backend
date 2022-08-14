package dao

import "github.com/teamyapp/cloud/app/entity"

type UserGroup interface {
	FindGroupByID(groupID uint64) (entity.UserGroup, error)
	FindAllGroups() ([]entity.UserGroup, error)
	CreateGroup(group entity.UserGroup) error
	UpdateGroup(group entity.UserGroup) error
	DeleteGroup(groupID uint64) error
}
