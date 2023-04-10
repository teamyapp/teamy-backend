package servicetest

import (
	"context"

	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/authorization"
)

func AddTeamPermission(
	ct context.Context,
	authorizationService service.Authorization,
	teamID uint64,
	groupID uint64,
	groupOperations []authorization.ResourceTypeOperation,
	requesterUserID uint64,
) *errs.Error {
	err := authorizationService.AddUserGroupMember(ct, groupID, requesterUserID)
	if err != nil {
		return err
	}

	for _, resourceTypeOperation := range groupOperations {
		err = authorizationService.RegisterOperation(
			ct,
			string(resourceTypeOperation.ResourceType),
			resourceTypeOperation.Operation)
		if err != nil {
			return err
		}

		err = authorizationService.AddPermission(
			ct,
			string(resourceTypeOperation.ResourceType),
			teamID,
			resourceTypeOperation.Operation,
			groupID)
		if err != nil {
			return err
		}
	}

	return authorizationService.RegisterResource(
		ct,
		string(authorization.TeamResourceType),
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
	groupOperations []authorization.ResourceTypeOperation,
	requesterUserID uint64,
) *errs.Error {
	err := authorizationService.AddUserGroupMember(ct, groupID, requesterUserID)
	if err != nil {
		return err
	}

	for _, resourceTypeOperation := range groupOperations {
		err = authorizationService.RegisterOperation(
			ct,
			string(resourceTypeOperation.ResourceType),
			resourceTypeOperation.Operation)
		if err != nil {
			return err
		}

		err = authorizationService.AddPermission(
			ct,
			string(resourceTypeOperation.ResourceType),
			appID,
			resourceTypeOperation.Operation,
			groupID)
		if err != nil {
			return err
		}
	}

	return authorizationService.RegisterResource(
		ct,
		string(authorization.AppResourceType),
		appID)
}
