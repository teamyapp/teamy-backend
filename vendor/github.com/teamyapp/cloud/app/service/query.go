package service

import (
	"strings"
	"time"

	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/libs/collect"
)

type ResourceTypeQuery struct {
	ResourceTypeName  *string
	CreatorUserID     *uint64
	StartCreationTime *time.Time
	EndCreationTime   *time.Time
	Limit             *uint64
}

type ResourceQuery struct {
	ResourceTypeName  *string
	ResourceID        *uint64
	CreatorUserID     *uint64
	StartCreationTime *time.Time
	EndCreationTime   *time.Time
	Limit             *uint64
}

type ResourceRelationQuery struct {
	ChildResourceType  *string
	ChildResourceID    *uint64
	ParentResourceType *string
	ParentResourceID   *uint64
	CreatorUserID      *uint64
	StartCreationTime  *time.Time
	EndCreationTime    *time.Time
	Limit              *uint64
}

type OperationQuery struct {
	ResourceTypeName  *string
	OperationName     *string
	CreatorUserID     *uint64
	StartCreationTime *time.Time
	EndCreationTime   *time.Time
	Limit             *uint64
}

type OperationRelationQuery struct {
	ChildResourceType  *string
	ChildOperation     *string
	ParentResourceType *string
	ParentOperation    *string
	CreatorUserID      *uint64
	StartCreationTime  *time.Time
	EndCreationTime    *time.Time
	Limit              *uint64
}

type UserGroupQuery struct {
	ID                  *uint64
	NameContains        *string
	DescriptionContains *string
	CreatorUserID       *uint64
	StartCreationTime   *time.Time
	EndCreationTime     *time.Time
	Limit               *uint64
}

func queryResourceTypes(resourceTypeEntities []entity.ResourceType, resourceTypeQuery ResourceTypeQuery) []entity.ResourceType {
	return collect.Filter(resourceTypeEntities, func(resourceTypeEntity entity.ResourceType) bool {
		if resourceTypeQuery.ResourceTypeName != nil && *resourceTypeQuery.ResourceTypeName != resourceTypeEntity.ResourceTypeName {
			return false
		}

		if resourceTypeQuery.CreatorUserID != nil && *resourceTypeQuery.CreatorUserID != resourceTypeEntity.CreatorUserID {
			return false
		}

		if resourceTypeQuery.StartCreationTime != nil && (*resourceTypeQuery.StartCreationTime).After(resourceTypeEntity.CreatedAt) {
			return false
		}

		if resourceTypeQuery.EndCreationTime != nil && (*resourceTypeQuery.EndCreationTime).Before(resourceTypeEntity.CreatedAt) {
			return false
		}

		return true
	})
}

func queryResources(resources []entity.Resource, resourceQuery ResourceQuery) []entity.Resource {
	return collect.Filter(resources, func(resource entity.Resource) bool {
		if resourceQuery.ResourceTypeName != nil && *resourceQuery.ResourceTypeName != resource.ResourceTypeName {
			return false
		}

		if resourceQuery.ResourceID != nil && *resourceQuery.ResourceID != resource.ResourceID {
			return false
		}

		if resourceQuery.CreatorUserID != nil && *resourceQuery.CreatorUserID != resource.CreatorUserID {
			return false
		}

		if resourceQuery.StartCreationTime != nil && (*resourceQuery.StartCreationTime).After(resource.CreatedAt) {
			return false
		}

		if resourceQuery.EndCreationTime != nil && (*resourceQuery.EndCreationTime).Before(resource.CreatedAt) {
			return false
		}

		return true
	})
}

func queryResourceRelations(resourceRelations []entity.ResourceRelation, resourceRelationQuery ResourceRelationQuery) []entity.ResourceRelation {
	return collect.Filter(resourceRelations, func(resourceRelation entity.ResourceRelation) bool {
		if resourceRelationQuery.ChildResourceID != nil && *resourceRelationQuery.ChildResourceID != resourceRelation.ChildResourceID {
			return false
		}

		if resourceRelationQuery.ChildResourceType != nil && *resourceRelationQuery.ChildResourceType != resourceRelation.ChildResourceType {
			return false
		}

		if resourceRelationQuery.ParentResourceID != nil && *resourceRelationQuery.ParentResourceID != resourceRelation.ParentResourceID {
			return false
		}

		if resourceRelationQuery.ParentResourceType != nil && *resourceRelationQuery.ParentResourceType != resourceRelation.ParentResourceType {
			return false
		}

		if resourceRelationQuery.CreatorUserID != nil && *resourceRelationQuery.CreatorUserID != resourceRelation.CreatorUserID {
			return false
		}

		if resourceRelationQuery.StartCreationTime != nil && (*resourceRelationQuery.StartCreationTime).After(resourceRelation.CreatedAt) {
			return false
		}

		if resourceRelationQuery.EndCreationTime != nil && (*resourceRelationQuery.EndCreationTime).Before(resourceRelation.CreatedAt) {
			return false
		}

		return true
	})
}

func queryOperations(operations []entity.Operation, operationQuery OperationQuery) []entity.Operation {
	return collect.Filter(operations, func(operation entity.Operation) bool {
		if operationQuery.ResourceTypeName != nil && *operationQuery.ResourceTypeName != operation.ResourceTypeName {
			return false
		}

		if operationQuery.OperationName != nil && *operationQuery.OperationName != operation.OperationName {
			return false
		}

		if operationQuery.CreatorUserID != nil && *operationQuery.CreatorUserID != operation.CreatorUserID {
			return false
		}

		if operationQuery.StartCreationTime != nil && (*operationQuery.StartCreationTime).After(operation.CreatedAt) {
			return false
		}

		if operationQuery.EndCreationTime != nil && (*operationQuery.EndCreationTime).Before(operation.CreatedAt) {
			return false
		}

		return true
	})
}

func queryOperationRelations(operationRelations []entity.OperationRelation, operationRelationQuery OperationRelationQuery) []entity.OperationRelation {
	return collect.Filter(operationRelations, func(operationRelation entity.OperationRelation) bool {
		if operationRelationQuery.ChildResourceType != nil && *operationRelationQuery.ChildResourceType != operationRelation.ChildResourceType {
			return false
		}

		if operationRelationQuery.ChildOperation != nil && *operationRelationQuery.ChildOperation != operationRelation.ChildOperation {
			return false
		}

		if operationRelationQuery.ParentResourceType != nil && *operationRelationQuery.ParentResourceType != operationRelation.ParentResourceType {
			return false
		}

		if operationRelationQuery.ParentOperation != nil && *operationRelationQuery.ParentOperation != operationRelation.ParentOperation {
			return false
		}

		if operationRelationQuery.CreatorUserID != nil && *operationRelationQuery.CreatorUserID != operationRelation.CreatorUserID {
			return false
		}

		if operationRelationQuery.StartCreationTime != nil && (*operationRelationQuery.StartCreationTime).After(operationRelation.CreatedAt) {
			return false
		}

		if operationRelationQuery.EndCreationTime != nil && (*operationRelationQuery.EndCreationTime).Before(operationRelation.CreatedAt) {
			return false
		}

		return true
	})
}

func queryUserGroups(userGroups []entity.UserGroup, userGroupQuery UserGroupQuery) []entity.UserGroup {
	return collect.Filter(userGroups, func(userGroup entity.UserGroup) bool {
		if userGroupQuery.ID != nil && *userGroupQuery.ID != userGroup.ID {
			return false
		}

		if userGroupQuery.NameContains != nil && !strings.Contains(userGroup.Name, *userGroupQuery.NameContains) {
			return false
		}

		if userGroupQuery.DescriptionContains != nil &&
			(userGroup.Description == nil || !strings.Contains(*userGroup.Description, *userGroupQuery.DescriptionContains)) {
			return false
		}

		if userGroupQuery.CreatorUserID != nil && *userGroupQuery.CreatorUserID != userGroup.CreatorUserID {
			return false
		}

		if userGroupQuery.StartCreationTime != nil && (*userGroupQuery.StartCreationTime).After(userGroup.CreatedAt) {
			return false
		}

		if userGroupQuery.EndCreationTime != nil && (*userGroupQuery.EndCreationTime).Before(userGroup.CreatedAt) {
			return false
		}

		return true
	})
}
