package service

import (
	"errors"
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type Authorization struct {
	resourceRelationDao  dao.ResourceRelation
	userGroupMemberDao   dao.UserGroupMember
	permissionDao        dao.Permission
	operationRelationDao dao.OperationRelation
}

func (a Authorization) HasPermission(resourceType string, resourceID uint64, operation string, userID uint64) (bool, error) {
	// No nested group allowed
	groupIDs, err := a.userGroupMemberDao.FindGroupIDsByUserID(userID)
	if err != nil {
		log.Println(err)
		return false, err
	}

	for _, groupID := range groupIDs {
		hasPermission, err := a.groupHasPermission(entity.PermissionQuery{
			ResourceID:   resourceID,
			ResourceType: resourceType,
			Operation:    operation,
			GroupID:      groupID,
		})
		if err != nil {
			log.Println(err)
			// Continue check permission in other groups if current group fails to grant permission
			continue
		}

		if hasPermission {
			return hasPermission, nil
		}
	}

	return false, err
}

func (a Authorization) groupHasPermission(permissionQuery entity.PermissionQuery) (bool, error) {
	visited := make(map[entity.PermissionQuery]bool)
	visited[permissionQuery] = true
	queries := []entity.PermissionQuery{permissionQuery}
	for len(queries) > 0 {
		currQuery := queries[0]
		queries = queries[1:]

		_, err := a.permissionDao.FindPermission(currQuery)
		if err == nil {
			return true, nil
		}

		var errNotFound dao.ErrNotFound
		if !errors.As(err, &errNotFound) {
			log.Println(err)
			continue
		}

		parentPermissionQueries, err := a.getParentPermissionQueries(currQuery, visited)
		if err != nil {
			return false, err
		}

		queries = append(queries, parentPermissionQueries...)
	}

	return false, nil
}

func (a Authorization) getParentPermissionQueries(currQuery entity.PermissionQuery, visited map[entity.PermissionQuery]bool) ([]entity.PermissionQuery, error) {
	var parentPermissionQueries []entity.PermissionQuery
	operationRelations, err := a.operationRelationDao.FindOperationRelations(currQuery.ResourceType, currQuery.Operation)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	resourceRelations, err := a.resourceRelationDao.FindResourceRelations(currQuery.ResourceID, currQuery.ResourceType)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for _, resourceRelation := range resourceRelations {
		if resourceRelation.ParentResourceType == currQuery.ResourceType {
			newPermissionQuery := entity.PermissionQuery{
				ResourceID:   resourceRelation.ParentResourceID,
				ResourceType: currQuery.ResourceType,
				Operation:    currQuery.Operation,
				GroupID:      currQuery.GroupID,
			}
			parentPermissionQueries = tryAddPermissionQueryToQueue(newPermissionQuery, visited, parentPermissionQueries)
		}
	}

	for _, operationRelation := range operationRelations {
		if operationRelation.ParentResourceType == currQuery.ResourceType {
			newPermissionQuery := entity.PermissionQuery{
				ResourceID:   currQuery.ResourceID,
				ResourceType: operationRelation.ParentResourceType,
				Operation:    operationRelation.ParentOperation,
				GroupID:      currQuery.GroupID,
			}
			parentPermissionQueries = tryAddPermissionQueryToQueue(newPermissionQuery, visited, parentPermissionQueries)
			continue
		}

		for _, resourceRelation := range resourceRelations {
			if resourceRelation.ParentResourceType != operationRelation.ParentResourceType {
				continue
			}

			newPermissionQuery := entity.PermissionQuery{
				ResourceID:   resourceRelation.ParentResourceID,
				ResourceType: operationRelation.ParentResourceType,
				Operation:    operationRelation.ParentOperation,
				GroupID:      currQuery.GroupID,
			}
			parentPermissionQueries = tryAddPermissionQueryToQueue(newPermissionQuery, visited, parentPermissionQueries)
		}
	}

	return parentPermissionQueries, nil
}

func tryAddPermissionQueryToQueue(permissionQuery entity.PermissionQuery, visited map[entity.PermissionQuery]bool, queries []entity.PermissionQuery) []entity.PermissionQuery {
	_, ok := visited[permissionQuery]
	if ok {
		return queries
	}

	visited[permissionQuery] = true
	return append(queries, permissionQuery)
}

func NewAuthorization(
	resourceRelationDao dao.ResourceRelation,
	userGroupMemberDao dao.UserGroupMember,
	permissionDao dao.Permission,
	operationRelationDao dao.OperationRelation,
) Authorization {
	return Authorization{
		resourceRelationDao:  resourceRelationDao,
		userGroupMemberDao:   userGroupMemberDao,
		permissionDao:        permissionDao,
		operationRelationDao: operationRelationDao,
	}
}
