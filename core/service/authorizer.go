package service

import (
	"context"
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

func newAuthorizer(
	dataCollector obs.DataCollector,
	cloudClientRegistry *api.ClientRegistry,
) Authorizer {
	return Authorizer{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
	}
}
