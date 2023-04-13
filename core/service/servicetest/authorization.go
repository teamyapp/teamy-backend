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
