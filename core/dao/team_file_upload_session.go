package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamFileUploadSession interface {
	FindTeamFileUploadSessionByTeamID(
		ct context.Context,
		teamID uint64,
		teamFileUploadSessionType entity.TeamFileUploadSessionType,
		fileUploadSessionID uint64,
	) (entity.TeamFileUploadSession, error)
	CreateTeamFileUploadSession(ct context.Context, teamFileUploadSession entity.TeamFileUploadSession) error
	UpdateTeamFileUploadSession(ct context.Context, teamFileUploadSession entity.TeamFileUploadSession) error
}
