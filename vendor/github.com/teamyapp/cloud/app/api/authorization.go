package api

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/entity"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/runner"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Authorization struct {
	authorizationService service.Authorization
	proto.UnimplementedAuthorizationServer
}

var _ runner.Service = (*Authorization)(nil)
var _ proto.AuthorizationServer = (*Authorization)(nil)

func (a Authorization) HasPermission(ctx context.Context, req *proto.HasPermissionRequest) (*proto.HasPermissionResponse, error) {
	hasPermission, err := a.authorizationService.HasPermission(req.ResourceType, req.ResourceId, req.Operation, req.UserId)
	if err != nil {
		return nil, err
	}

	return &proto.HasPermissionResponse{HasPermission: hasPermission}, nil
}

func (a Authorization) ListResourceTypes(ct context.Context, query *proto.ListResourceTypesQuery) (*proto.ListResourceTypesResponse, error) {
	resourceTypeQuery := service.ResourceTypeQuery{
		ResourceTypeName: query.ResourceType,
		CreatorUserID:    query.CreatorUserId,
		Limit:            query.Limit,
	}

	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		resourceTypeQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		resourceTypeQuery.EndCreationTime = &endCreationTime
	}

	resourceTypeEntities, err := a.authorizationService.ListResourceTypes(ct, resourceTypeQuery)
	if err != nil {
		return nil, err
	}

	var resourceTypes []*proto.ResourceType
	resourceTypes = collect.Map(resourceTypeEntities, func(resourceTypeEntity entity.ResourceType, _ int) *proto.ResourceType {
		return &proto.ResourceType{
			ResourceType:  resourceTypeEntity.ResourceTypeName,
			CreatedAt:     timestamppb.New(resourceTypeEntity.CreatedAt),
			CreatorUserId: resourceTypeEntity.CreatorUserID,
		}
	})
	return &proto.ListResourceTypesResponse{ResourceTypes: resourceTypes}, nil
}

func (a Authorization) RegisterResourceType(ct context.Context, request *proto.RegisterResourceTypeRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RegisterResourceType(ct, request.ResourceType)
	return &emptypb.Empty{}, err
}

func (a Authorization) UnregisterResourceType(ct context.Context, request *proto.UnregisterResourceTypeRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnregisterResourceType(ct, request.ResourceType)
	return &emptypb.Empty{}, err
}

func (a Authorization) ListResources(ct context.Context, query *proto.ListResourcesQuery) (*proto.ListResourcesResponse, error) {
	resourceQuery := service.ResourceQuery{
		ResourceTypeName: query.ResourceType,
		ResourceID:       query.ResourceId,
		CreatorUserID:    query.CreatorUserId,
		Limit:            query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		resourceQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		resourceQuery.EndCreationTime = &endCreationTime
	}

	resourceEntities, err := a.authorizationService.ListResources(ct, resourceQuery)
	if err != nil {
		return nil, err
	}

	resources := collect.Map(resourceEntities, func(resource entity.Resource, _ int) *proto.Resource {
		return &proto.Resource{
			ResourceType:  resource.ResourceTypeName,
			ResourceId:    resource.ResourceID,
			CreatedAt:     timestamppb.New(resource.CreatedAt),
			CreatorUserId: resource.CreatorUserID,
		}
	})

	return &proto.ListResourcesResponse{Resources: resources}, nil
}

func (a Authorization) RegisterResource(ct context.Context, request *proto.RegisterResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RegisterResource(ct, request.ResourceType, request.ResourceId)
	return &emptypb.Empty{}, err
}

func (a Authorization) UnregisterResource(ct context.Context, request *proto.UnregisterResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnregisterResource(ct, request.ResourceType, request.ResourceId)
	return &emptypb.Empty{}, err
}

func (a Authorization) ListResourceRelations(ct context.Context, query *proto.ListResourceRelationsQuery) (*proto.ListResourceRelationsResponse, error) {
	resourceRelationQuery := service.ResourceRelationQuery{
		ChildResourceType:  query.ChildResourceType,
		ChildResourceID:    query.ChildResourceId,
		ParentResourceType: query.ParentResourceType,
		ParentResourceID:   query.ParentResourceId,
		CreatorUserID:      query.CreatorUserId,
		Limit:              query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		resourceRelationQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		resourceRelationQuery.EndCreationTime = &endCreationTime
	}

	resourceRelationEntities, err := a.authorizationService.ListResourceRelations(ct, resourceRelationQuery)
	if err != nil {
		return nil, err
	}

	resourceRelations := collect.Map(resourceRelationEntities, func(resourceRelation entity.ResourceRelation, _ int) *proto.ResourceRelation {
		return &proto.ResourceRelation{
			ChildResourceType:  resourceRelation.ChildResourceType,
			ChildResourceId:    resourceRelation.ChildResourceID,
			ParentResourceType: resourceRelation.ParentResourceType,
			ParentResourceId:   resourceRelation.ParentResourceID,
			CreatedAt:          timestamppb.New(resourceRelation.CreatedAt),
			CreatorUserId:      resourceRelation.CreatorUserID,
		}
	})

	return &proto.ListResourceRelationsResponse{ResourceRelations: resourceRelations}, nil
}

func (a Authorization) AssignParentResource(ct context.Context, request *proto.AssignParentResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.AssignParentResource(
		ct,
		request.ChildResourceType,
		request.ChildResourceId,
		request.ParentResourceType,
		request.ParentResourceId,
	)
	return &emptypb.Empty{}, err
}

func (a Authorization) UnassignParentResource(ct context.Context, request *proto.UnassignParentResourceRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnassignParentResource(
		ct,
		request.ChildResourceType,
		request.ChildResourceId,
		request.ParentResourceType,
		request.ParentResourceId,
	)
	return &emptypb.Empty{}, err
}

func (a Authorization) ListOperations(ct context.Context, query *proto.ListOperationsQuery) (*proto.ListOperationsResponse, error) {
	operationQuery := service.OperationQuery{
		ResourceTypeName: query.ResourceType,
		OperationName:    query.Operation,
		CreatorUserID:    query.CreatorUserId,
		Limit:            query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		operationQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		operationQuery.EndCreationTime = &endCreationTime
	}

	operationEntities, err := a.authorizationService.ListOperations(ct, operationQuery)
	if err != nil {
		return nil, err
	}

	operations := collect.Map(operationEntities, func(operation entity.Operation, _ int) *proto.Operation {
		return &proto.Operation{
			ResourceType:  operation.ResourceTypeName,
			Operation:     operation.OperationName,
			CreatedAt:     timestamppb.New(operation.CreatedAt),
			CreatorUserId: operation.CreatorUserID,
		}
	})

	return &proto.ListOperationsResponse{Operations: operations}, nil
}

func (a Authorization) RegisterOperation(ct context.Context, request *proto.RegisterOperationRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.RegisterOperation(ct, request.ResourceType, request.Operation)
	return &emptypb.Empty{}, err
}

func (a Authorization) UnregisterOperation(ct context.Context, request *proto.UnregisterOperationRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnregisterOperation(ct, request.ResourceType, request.Operation)
	return &emptypb.Empty{}, err
}

func (a Authorization) ListOperationRelations(ct context.Context, query *proto.ListOperationRelationsQuery) (*proto.ListOperationRelationsResponse, error) {
	operationRelationQuery := service.OperationRelationQuery{
		ChildResourceType:  query.ChildResourceType,
		ChildOperation:     query.ChildOperation,
		ParentResourceType: query.ParentResourceType,
		ParentOperation:    query.ParentOperation,
		CreatorUserID:      query.CreatorUserId,
		Limit:              query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		operationRelationQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		operationRelationQuery.EndCreationTime = &endCreationTime
	}

	operationRelationEntities, err := a.authorizationService.ListOperationRelations(ct, operationRelationQuery)
	if err != nil {
		return nil, err
	}

	operationRelations := collect.Map(operationRelationEntities, func(operationRelation entity.OperationRelation, _ int) *proto.OperationRelation {
		return &proto.OperationRelation{
			ChildResourceType:  operationRelation.ChildResourceType,
			ChildOperation:     operationRelation.ChildOperation,
			ParentResourceType: operationRelation.ParentResourceType,
			ParentOperation:    operationRelation.ParentOperation,
			CreatedAt:          timestamppb.New(operationRelation.CreatedAt),
			CreatorUserId:      operationRelation.CreatorUserID,
		}
	})

	return &proto.ListOperationRelationsResponse{OperationRelations: operationRelations}, nil
}

func (a Authorization) AssignParentOperation(ct context.Context, request *proto.AssignParentOperationRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.AssignParentOperation(
		ct,
		request.ChildResourceType,
		request.ChildOperation,
		request.ParentResourceType,
		request.ParentOperation,
	)
	return &emptypb.Empty{}, err
}

func (a Authorization) UnassignParentOperation(ct context.Context, request *proto.UnassignParentOperationRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UnassignParentOperation(
		ct,
		request.ChildResourceType,
		request.ChildOperation,
		request.ParentResourceType,
		request.ParentOperation,
	)
	return &emptypb.Empty{}, err
}

func (a Authorization) ListUserGroups(ct context.Context, query *proto.ListUserGroupsQuery) (*proto.ListUserGroupsResponse, error) {
	userGroupQuery := service.UserGroupQuery{
		ID:                  query.Id,
		NameContains:        query.NameContains,
		DescriptionContains: query.DescriptionContains,
		CreatorUserID:       query.CreatorUserId,
		Limit:               query.Limit,
	}
	if query.StartCreationTime != nil {
		startCreationTime := query.StartCreationTime.AsTime()
		userGroupQuery.StartCreationTime = &startCreationTime
	}

	if query.EndCreationTime != nil {
		endCreationTime := query.EndCreationTime.AsTime()
		userGroupQuery.EndCreationTime = &endCreationTime
	}

	userGroupEntities, err := a.authorizationService.ListUserGroups(ct, userGroupQuery)
	if err != nil {
		return nil, err
	}

	userGroups := collect.Map(userGroupEntities, func(userGroup entity.UserGroup, _ int) *proto.UserGroup {
		return &proto.UserGroup{
			GroupId:       userGroup.ID,
			Name:          userGroup.Name,
			Description:   userGroup.Description,
			CreatedAt:     timestamppb.New(userGroup.CreatedAt),
			CreatorUserId: userGroup.CreatorUserID,
			UpdatedAt:     toProtoTimePtr(userGroup.UpdatedAt),
		}
	})
	return &proto.ListUserGroupsResponse{UserGroups: userGroups}, nil
}

func (a Authorization) CreateUserGroup(ct context.Context, request *proto.CreateUserGroupRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.CreateUserGroup(ct, request.Name, request.Description)
	return &emptypb.Empty{}, err
}

func (a Authorization) UpdateUserGroup(ct context.Context, request *proto.UpdateUserGroupRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.UpdateUserGroup(ct, request.GroupId, request.Name, request.Description)
	return &emptypb.Empty{}, err
}

func (a Authorization) DeleteUserGroup(ct context.Context, request *proto.DeleteUserGroupRequest) (*emptypb.Empty, error) {
	err := a.authorizationService.DeleteUserGroup(ct, request.GroupId)
	return &emptypb.Empty{}, err
}

func (a Authorization) ListUserGroupMembers(ctx context.Context, query *proto.ListUserGroupMembersQuery) (*proto.ListUserGroupMembersResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) AddUserGroupMember(ctx context.Context, request *proto.AddUserGroupMemberRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) RemoveUserGroupMember(ctx context.Context, request *proto.RemoveUserGroupMemberRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) ListPermissions(ctx context.Context, query *proto.ListPermissionsQuery) (*proto.ListPermissionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) AddPermission(ctx context.Context, request *proto.AddPermissionRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) RemovePermission(ctx context.Context, request *proto.RemovePermissionRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (a Authorization) Start(rn *runner.ServiceRunner) error {
	rn.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterAuthorizationServer(server, a)
	})
	return nil
}

func NewAuthorization(authorizationService service.Authorization) Authorization {
	return Authorization{
		authorizationService: authorizationService,
	}
}
