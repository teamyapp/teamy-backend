package gql

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/authorization"
)

func (m Mutation) hasPermission(ct context.Context, query authorization.Query) (bool, error) {
	hasPermissionReq := &proto.HasPermissionRequest{
		ResourceType: string(query.ResourceType),
		ResourceId:   query.ResourceID,
		Operation:    string(query.Operation),
		UserId:       query.UserID,
	}
	hasPermissionRes, err := m.deps.cloudClientRegistry.AuthorizationClient().HasPermission(ct, hasPermissionReq)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return false, err
	}

	return hasPermissionRes.HasPermission, nil
}
