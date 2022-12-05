package service

import (
	"context"

	"fmt"
	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/authorization"
)

type Authorizer struct {
	dataCollector       obs.DataCollector
	cloudClientRegistry *api.ClientRegistry
}

func (a Authorizer) hasPermission(ct context.Context, query authorization.Query) (bool, error) {
	hasPermissionReq := &proto.HasPermissionRequest{
		ResourceType: string(query.ResourceType),
		ResourceId:   query.ResourceID,
		Operation:    query.Operation,
		UserId:       query.UserID,
	}
	hasPermissionRes, err := a.cloudClientRegistry.AuthorizationClient().HasPermission(ct, hasPermissionReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return false, err
	}

	return hasPermissionRes.HasPermission, nil
}

func (a Authorizer) registerResource(ct context.Context, resourceType authorization.ResourceType, resourceID uint64) error {
	registerResourceReq := &proto.RegisterResourceRequest{
		ResourceType: string(resourceType),
		ResourceId:   resourceID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().RegisterResource(ct, registerResourceReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (a Authorizer) assignParentResource(
	ct context.Context,
	childResourceType authorization.ResourceType,
	childResourceID uint64,
	parentResourceType authorization.ResourceType,
	parentResourceID uint64) error {
	assignParentResourceReq := &proto.AssignParentResourceRequest{
		ChildResourceType:  string(childResourceType),
		ChildResourceId:    childResourceID,
		ParentResourceType: string(parentResourceType),
		ParentResourceId:   parentResourceID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().AssignParentResource(ct, assignParentResourceReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (a Authorizer) addMemberToUserGroup(ct context.Context, userGroupID uint64, memberID uint64) error {
	addUserGroupMemberReq := &proto.AddUserGroupMemberRequest{
		GroupId: userGroupID,
		UserId:  memberID,
	}

	_, err := a.cloudClientRegistry.AuthorizationClient().AddUserGroupMember(ct, addUserGroupMemberReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (a Authorizer) createUserGroup(ct context.Context, creatorUserID uint64, userGroupName string, description *string) (uint64, error) {
	createUserGroupReq := &proto.CreateUserGroupRequest{
		Name:        userGroupName,
		Description: description,
	}

	createUserGroupRes, err := a.cloudClientRegistry.AuthorizationClient().CreateUserGroup(ct, createUserGroupReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	// add the group creator to the newly created userGroup
	err = a.addMemberToUserGroup(ct, createUserGroupRes.UserGroup.GroupId, creatorUserID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	a.dataCollector.Logger.LogWithContext(ct, obs.Info,
		obs.Props{obs.MessageProp: fmt.Sprintf("UserGroup %s is successfully created", userGroupName)},
	)

	return createUserGroupRes.UserGroup.GroupId, nil
}

func (a Authorizer) assignPermission(
	ct context.Context,
	resourceType authorization.ResourceType,
	resourceID uint64,
	operation authorization.Operation,
	userGroupID uint64,
) error {
	addPermissionReq := &proto.AddPermissionRequest{
		ResourceType: string(resourceType),
		ResourceId:   resourceID,
		Operation:    string(operation),
		GroupId:      userGroupID,
	}

	_, err := a.cloudClientRegistry.AuthorizationClient().AddPermission(ct, addPermissionReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	a.dataCollector.Logger.LogWithContext(ct, obs.Info,
		obs.Props{obs.MessageProp: fmt.Sprintf("Permission %s is successfully assigned", addPermissionReq)},
	)

	return nil
}

func (a Authorizer) assignUserGroupPermissions(
	ct context.Context,
	resourceType authorization.ResourceType,
	resourceID uint64,
	groupID uint64,
	userGroupOperations []authorization.Operation,
) error {
	for _, operation := range userGroupOperations {
		err := a.assignPermission(ct, resourceType, resourceID, operation, groupID)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}
	}

	return nil
}

func (a Authorizer) createUserGroupAndAssignPermissions(
	ct context.Context,
	creatorUserID uint64,
	userGroupName string,
	description *string,
	resourceType authorization.ResourceType,
	resourceID uint64,
	userGroupOperations []authorization.Operation,
) (uint64, error) {
	userGroupID, err := a.createUserGroup(ct, creatorUserID, userGroupName, description)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	err = a.assignUserGroupPermissions(ct, resourceType, resourceID, userGroupID, userGroupOperations)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return 0, err
	}

	return userGroupID, nil
}

func NewAuthorizer(
	dataCollector obs.DataCollector,
	cloudClientRegistry *api.ClientRegistry,
) Authorizer {
	return Authorizer{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
	}
}
