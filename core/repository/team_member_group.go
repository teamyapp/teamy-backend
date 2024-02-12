package repository

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	daoEntity "github.com/teamyapp/teamy-backend/core/dao/entity"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMemberGroup struct {
	teamMemberGroupDao             dao.TeamMemberGroup
	teamMemberGroupUserRelationDao dao.TeamMemberGroupUserRelation
}

func (t TeamMemberGroup) FindMemberGroupByID(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
) (entity.TeamMemberGroup, *errs.Error) {
	teamMemberGroup, err := t.teamMemberGroupDao.FindMemberGroupByID(ct, tx, groupID)
	if err != nil {
		return entity.TeamMemberGroup{}, err
	}

	memberUserIDs, err := t.teamMemberGroupUserRelationDao.FindMemberGroupUserIDsByMemberGroupID(ct, tx, groupID)
	if err != nil {
		return entity.TeamMemberGroup{}, err
	}

	return entity.TeamMemberGroup{
		ID:                       teamMemberGroup.ID,
		Name:                     teamMemberGroup.Name,
		TeamID:                   teamMemberGroup.TeamID,
		AuthorizationUserGroupID: teamMemberGroup.AuthorizationUserGroupID,
		MemberUserIDs:            memberUserIDs,
		CreatedAt:                teamMemberGroup.CreatedAt,
		UpdatedAt:                teamMemberGroup.UpdatedAt,
	}, nil
}

func (t TeamMemberGroup) FindMemberGroupsByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamMemberGroup, *errs.Error) {
	rawTeamMemberGroups, err := t.teamMemberGroupDao.FindMemberGroupsByTeamID(ct, tx, teamID)
	if err != nil {
		return nil, err
	}

	var teamMemberGroups []entity.TeamMemberGroup
	for _, rawTeamMemberGroup := range rawTeamMemberGroups {
		memberUserIDs, err := t.teamMemberGroupUserRelationDao.FindMemberGroupUserIDsByMemberGroupID(ct, tx, rawTeamMemberGroup.ID)
		if err != nil {
			return nil, err
		}

		teamMemberGroups = append(teamMemberGroups, entity.TeamMemberGroup{
			ID:                       rawTeamMemberGroup.ID,
			Name:                     rawTeamMemberGroup.Name,
			TeamID:                   rawTeamMemberGroup.TeamID,
			AuthorizationUserGroupID: rawTeamMemberGroup.AuthorizationUserGroupID,
			MemberUserIDs:            memberUserIDs,
			CreatedAt:                rawTeamMemberGroup.CreatedAt,
			UpdatedAt:                rawTeamMemberGroup.UpdatedAt,
		})
	}

	return teamMemberGroups, nil
}

func GetTeamMemberGroupFromRawTeamMemberGroup(teamMemberGroup daoEntity.TeamMemberGroup) entity.TeamMemberGroup {
	return entity.TeamMemberGroup{
		ID:                       teamMemberGroup.ID,
		TeamID:                   teamMemberGroup.TeamID,
		Name:                     teamMemberGroup.Name,
		AuthorizationUserGroupID: teamMemberGroup.AuthorizationUserGroupID,
		MemberUserIDs:            []uint64{},
		CreatedAt:                teamMemberGroup.CreatedAt,
		UpdatedAt:                teamMemberGroup.UpdatedAt,
	}
}

func NewTeamMemberGroup(
	teamMemberGroupDao dao.TeamMemberGroup,
	teamMemberGroupUserRelationDao dao.TeamMemberGroupUserRelation,
) TeamMemberGroup {
	return TeamMemberGroup{
		teamMemberGroupDao:             teamMemberGroupDao,
		teamMemberGroupUserRelationDao: teamMemberGroupUserRelationDao,
	}
}
