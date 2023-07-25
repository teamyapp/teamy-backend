package servicetest

import (
	"context"

	"github.com/teamyapp/cloud/app/service"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core"
	"github.com/teamyapp/teamy-backend/core/authorization"
)

func AddTeamPermission(
	ct context.Context,
	authorizationService service.Authorization,
	teamID uint64,
	groupID uint64,
	groupOperations []cloudAuthorization.ResourceTypeOperation,
	requesterUserID uint64,
) *errs.Error {
	err := authorizationService.AddUserGroupMember(ct, groupID, requesterUserID)
	if err != nil {
		return err
	}

	for _, resourceTypeOperation := range groupOperations {
		err = authorizationService.AddPermission(
			ct,
			resourceTypeOperation.ResourceType,
			teamID,
			resourceTypeOperation.Operation,
			groupID)
		if err != nil {
			return err
		}
	}

	return authorizationService.RegisterResource(
		ct,
		authorization.TeamResourceType,
		teamID)
}

func GetServiceAccountAPIToken(identityService service.Identity) (string, *errs.Error) {
	ct := context.Background()
	var accountOwner uint64 = 0
	serviceAccountID, internalErr := identityService.CreateServiceAccount(ct, accountOwner, "test")
	if internalErr != nil {
		return "", internalErr
	}

	apiToken, internalErr := identityService.GenerateServiceToken(ct, accountOwner, serviceAccountID)
	if internalErr != nil {
		return "", internalErr
	}

	return apiToken, nil
}

func AddAppPermission(
	ct context.Context,
	authorizationService service.Authorization,
	appID uint64,
	groupID uint64,
	groupOperations []cloudAuthorization.ResourceTypeOperation,
	requesterUserID uint64,
) *errs.Error {
	err := authorizationService.AddUserGroupMember(ct, groupID, requesterUserID)
	if err != nil {
		return err
	}

	for _, resourceTypeOperation := range groupOperations {
		err = authorizationService.AddPermission(
			ct,
			resourceTypeOperation.ResourceType,
			appID,
			resourceTypeOperation.Operation,
			groupID)
		if err != nil {
			return err
		}
	}

	return authorizationService.RegisterResource(
		ct,
		authorization.AppResourceType,
		appID)
}

func AddAllTeamPermissions(
	authorizationService service.Authorization,
	teamID uint64,
	userID uint64,
) *errs.Error {
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, 1)
	group, err := authorizationService.
		CreateUserGroup(ct, "FullControl", nil)
	if err != nil {
		return err
	}

	return AddTeamPermission(
		ct,
		authorizationService,
		group.ID,
		teamID,
		authorization.AllTeamResourceTypeOperations,
		userID)
}

func ApplyAuthorizationConfig(authorizationService service.Authorization) *errs.Error {
	ct := context.Background()
	ct = ctx.NewContextWithUserID(ct, 1)
	return authorizationService.ApplyAuthorizationConfig(ct, core.AuthorizationConfig)
}
