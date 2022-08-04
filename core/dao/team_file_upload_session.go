package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamFileUploadSession interface {
	FindTeamFileUploadSessionByTeamID(
		teamID uint64,
		teamFileUploadSessionType entity.TeamFileUploadSessionType,
		fileUploadSessionID uint64,
	) (entity.TeamFileUploadSession, error)
	CreateTeamFileUploadSession(teamFileUploadSession entity.TeamFileUploadSession) error
	UpdateTeamFileUploadSession(teamFileUploadSession entity.TeamFileUploadSession) error
}
