package service

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/authorization"
)

type Authorizer struct {
	dataCollector       telemetry.DataCollector
	cloudClientRegistry *api.ClientRegistry
}

func (a Authorizer) hasPermission(ct context.Context, query authorization.Query) (bool, *errs.Error) {
	hasPermissionReq := &proto.HasPermissionRequest{
		ResourceType: string(query.ResourceType),
		ResourceId:   query.ResourceID,
		Operation:    query.Operation,
		UserId:       query.UserID,
	}
	hasPermissionRes, err := a.cloudClientRegistry.AuthorizationClient().HasPermission(ct, hasPermissionReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return false, errs.FromGRPCErr(err)
	}

	return hasPermissionRes.HasPermission, nil
}

func (a Authorizer) registerResource(ct context.Context, resourceType authorization.ResourceType, resourceID uint64) *errs.Error {
	registerResourceReq := &proto.RegisterResourceRequest{
		ResourceType: string(resourceType),
		ResourceId:   resourceID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().RegisterResource(ct, registerResourceReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return errs.FromGRPCErr(err)
	}

	return nil
}

func (a Authorizer) assignParentResource(
	ct context.Context,
	childResourceType authorization.ResourceType,
	childResourceID uint64,
	parentResourceType authorization.ResourceType,
	parentResourceID uint64) *errs.Error {
	assignParentResourceReq := &proto.AssignParentResourceRequest{
		ChildResourceType:  string(childResourceType),
		ChildResourceId:    childResourceID,
		ParentResourceType: string(parentResourceType),
		ParentResourceId:   parentResourceID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().AssignParentResource(ct, assignParentResourceReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return errs.FromGRPCErr(err)
	}

	return nil
}

func (a Authorizer) addMemberToUserGroup(ct context.Context, userGroupID uint64, memberID uint64) *errs.Error {
	addUserGroupMemberReq := &proto.AddUserGroupMemberRequest{
		GroupId: userGroupID,
		UserId:  memberID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().AddUserGroupMember(ct, addUserGroupMemberReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return errs.FromGRPCErr(err)
	}

	return nil
}

func (a Authorizer) createUserGroup(ct context.Context, creatorUserID uint64, userGroupName string, description *string) (uint64, *errs.Error) {
	createUserGroupReq := &proto.CreateUserGroupRequest{
		Name:        userGroupName,
		Description: description,
	}

	createUserGroupRes, err := a.cloudClientRegistry.AuthorizationClient().CreateUserGroup(ct, createUserGroupReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return 0, errs.FromGRPCErr(err)
	}

	// add the group creator to the newly created userGroup
	err = a.addMemberToUserGroup(ct, createUserGroupRes.UserGroup.GroupId, creatorUserID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return 0, errs.FromGRPCErr(err)
	}

	a.dataCollector.Logger.LogWithContext(ct, telemetry.Info,
		telemetry.Props{
			telemetry.MessageProp: fmt.Sprintf("UserGroup %s is successfully created", userGroupName),
		},
	)

	return createUserGroupRes.UserGroup.GroupId, nil
}

func (a Authorizer) assignPermission(
	ct context.Context,
	resourceOperation authorization.ResourceOperation,
	userGroupID uint64,
) *errs.Error {
	addPermissionReq := &proto.AddPermissionRequest{
		ResourceType: string(resourceOperation.ResourceType),
		ResourceId:   resourceOperation.ResourceID,
		Operation:    resourceOperation.Operation,
		GroupId:      userGroupID,
	}
	_, err := a.cloudClientRegistry.AuthorizationClient().AddPermission(ct, addPermissionReq)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return errs.FromGRPCErr(err)
	}

	a.dataCollector.Logger.LogWithContext(ct, telemetry.Info,
		telemetry.Props{
			telemetry.MessageProp: fmt.Sprintf("Permission %s is successfully assigned", addPermissionReq),
		},
	)

	return nil
}

func (a Authorizer) assignUserGroupPermissions(
	ct context.Context,
	resourceOperations []authorization.ResourceOperation,
	groupID uint64,
) *errs.Error {
	for _, resourceOperation := range resourceOperations {
		err := a.assignPermission(ct, resourceOperation, groupID)
		if err != nil {
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
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
	resourceOperations []authorization.ResourceOperation,
) (uint64, *errs.Error) {
	userGroupID, err := a.createUserGroup(ct, creatorUserID, userGroupName, description)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return 0, err
	}

	err = a.assignUserGroupPermissions(ct, resourceOperations, userGroupID)
	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return 0, err
	}

	return userGroupID, nil
}

func NewAuthorizer(
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *api.ClientRegistry,
) Authorizer {
	return Authorizer{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
	}
}
