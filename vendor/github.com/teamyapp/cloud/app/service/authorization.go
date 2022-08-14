package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/teamyapp/cloud/app/gen"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/ctx"
)

type Authorization struct {
	resourceRelationDao  dao.ResourceRelation
	userGroupMemberDao   dao.UserGroupMember
	permissionDao        dao.Permission
	operationRelationDao dao.OperationRelation
	operationDao         dao.Operation
	resourceTypeDao      dao.ResourceType
	resourceDao          dao.Resource
	userGroupDao         dao.UserGroup
	userGroupIDGenerator *gen.UniqueNumber
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

func (a Authorization) ListResourceTypes(ct context.Context, resourceTypeQuery ResourceTypeQuery) ([]entity.ResourceType, error) {
	allResourceTypeEntities, err := a.resourceTypeDao.FindAllResourceTypes()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return queryResourceTypes(allResourceTypeEntities, resourceTypeQuery), nil
}

func (a Authorization) RegisterResourceType(ct context.Context, resourceTypeName string) error {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return err
	}

	resourceTypeEntity := entity.ResourceType{
		ResourceTypeName: resourceTypeName,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}

	return a.resourceTypeDao.CreateResourceType(resourceTypeEntity)
}

func (a Authorization) UnregisterResourceType(ct context.Context, resourceTypeName string) error {
	return a.resourceTypeDao.DeleteResourceType(resourceTypeName)
}

func (a Authorization) ListResources(ct context.Context, resourceQuery ResourceQuery) ([]entity.Resource, error) {
	allResources, err := a.resourceDao.FindAllResources()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return queryResources(allResources, resourceQuery), nil
}

func (a Authorization) RegisterResource(ct context.Context, resourceTypeName string, resourceID uint64) error {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return err
	}

	resource := entity.Resource{
		ResourceTypeName: resourceTypeName,
		ResourceID:       resourceID,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}
	return a.resourceDao.CreateResource(resource)
}

func (a Authorization) UnregisterResource(ct context.Context, resourceTypeName string, resourceID uint64) error {
	return a.resourceDao.DeleteResource(resourceTypeName, resourceID)
}

func (a Authorization) ListResourceRelations(ct context.Context, resourceRelationQuery ResourceRelationQuery) ([]entity.ResourceRelation, error) {
	allResourceRelationEntities, err := a.resourceRelationDao.FindAllResourceRelations()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return queryResourceRelations(allResourceRelationEntities, resourceRelationQuery), nil
}

func (a Authorization) AssignParentResource(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) error {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return err
	}

	resourceRelation := entity.ResourceRelation{
		ChildResourceType:  childResourceType,
		ChildResourceID:    childResourceID,
		ParentResourceType: parentResourceType,
		ParentResourceID:   parentResourceID,
		CreatedAt:          time.Now().UTC(),
		CreatorUserID:      userID,
	}
	return a.resourceRelationDao.CreateResourceRelation(resourceRelation)
}

func (a Authorization) UnassignParentResource(
	ct context.Context,
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) error {
	return a.resourceRelationDao.DeleteResourceRelation(
		childResourceType,
		childResourceID,
		parentResourceType,
		parentResourceID,
	)
}

func (a Authorization) ListOperations(ct context.Context, operationQuery OperationQuery) ([]entity.Operation, error) {
	allOperations, err := a.operationDao.FindAllOperations()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return queryOperations(allOperations, operationQuery), nil
}

func (a Authorization) RegisterOperation(ct context.Context, resourceTypeName string, operationName string) error {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return err
	}

	operation := entity.Operation{
		ResourceTypeName: resourceTypeName,
		OperationName:    operationName,
		CreatedAt:        time.Now().UTC(),
		CreatorUserID:    userID,
	}
	return a.operationDao.CreateOperation(operation)
}

func (a Authorization) UnregisterOperation(ct context.Context, resourceTypeName string, operationName string) error {
	return a.operationDao.DeleteOperation(resourceTypeName, operationName)
}

func (a Authorization) ListOperationRelations(ct context.Context, operationRelationQuery OperationRelationQuery) ([]entity.OperationRelation, error) {
	allOperationRelations, err := a.operationRelationDao.FindAllOperationRelations()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return queryOperationRelations(allOperationRelations, operationRelationQuery), nil
}

func (a Authorization) AssignParentOperation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) error {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return err
	}

	operationRelation := entity.OperationRelation{
		ChildResourceType:  childResourceType,
		ChildOperation:     childOperation,
		ParentResourceType: parentResourceType,
		ParentOperation:    parentOperation,
		CreatedAt:          time.Now().UTC(),
		CreatorUserID:      userID,
	}
	return a.operationRelationDao.CreateOperationRelation(operationRelation)
}

func (a Authorization) UnassignParentOperation(
	ct context.Context,
	childResourceType string,
	childOperation string,
	parentResourceType string,
	parentOperation string,
) error {
	return a.operationRelationDao.DeleteOperationRelation(
		childResourceType,
		childOperation,
		parentResourceType,
		parentOperation,
	)
}

func (a Authorization) ListUserGroups(ct context.Context, query UserGroupQuery) ([]entity.UserGroup, error) {
	allGroups, err := a.userGroupDao.FindAllGroups()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return queryUserGroups(allGroups, query), nil
}

func (a Authorization) CreateUserGroup(ct context.Context, name string, description *string) error {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return err
	}

	groupID, err := a.userGroupIDGenerator.GenerateUniqueNumber()
	if err != nil {
		log.Println(err)
		return err
	}

	userGroup := entity.UserGroup{
		ID:            groupID,
		Name:          name,
		Description:   description,
		CreatedAt:     time.Now().UTC(),
		CreatorUserID: userID,
	}
	return a.userGroupDao.CreateGroup(userGroup)
}

func (a Authorization) UpdateUserGroup(ct context.Context, groupID uint64, name *string, description *string) error {
	userGroup, err := a.userGroupDao.FindGroupByID(groupID)
	if err != nil {
		log.Println(err)
		return err
	}

	if name != nil {
		userGroup.Name = *name
	}

	if description != nil {
		userGroup.Description = description
	}

	nowTime := time.Now().UTC()
	userGroup.UpdatedAt = &nowTime
	return a.userGroupDao.UpdateGroup(userGroup)
}

func (a Authorization) DeleteUserGroup(ct context.Context, groupID uint64) error {
	return a.userGroupDao.DeleteGroup(groupID)
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

	resourceRelations, err := a.resourceRelationDao.FindResourceRelations(currQuery.ResourceType, currQuery.ResourceID)
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
	operationDao dao.Operation,
	resourceTypeDao dao.ResourceType,
	resourceDao dao.Resource,
	userGroupDao dao.UserGroup,
	uniqueNumberFactory gen.UniqueNumberFactory,
) (Authorization, error) {
	userGroupIDGenerator, err := uniqueNumberFactory.MakeUniqueNumber("userGroupID")
	if err != nil {
		return Authorization{}, err
	}

	return Authorization{
		resourceRelationDao:  resourceRelationDao,
		userGroupMemberDao:   userGroupMemberDao,
		permissionDao:        permissionDao,
		operationRelationDao: operationRelationDao,
		operationDao:         operationDao,
		resourceTypeDao:      resourceTypeDao,
		resourceDao:          resourceDao,
		userGroupDao:         userGroupDao,
		userGroupIDGenerator: userGroupIDGenerator,
	}, nil
}
