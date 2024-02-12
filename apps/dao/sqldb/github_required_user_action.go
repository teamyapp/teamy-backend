package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubRequiredUserAction struct {
	logger telemetry.Logger
	db     *sql.DB
}

var _ dao.GithubRequiredUserAction = (*GithubRequiredUserAction)(nil)

func (g GithubRequiredUserAction) FindRequiredUserActionsByActionUserID(
	ct context.Context,
	teamID uint64,
	actionUserID uint64,
) ([]entity.GithubRequiredUserAction, *errs.Error) {
	rows, err := g.db.Query(`
	SELECT
	    team_id,
	    action_user_id,
	    user_action_type,
	    is_completed,
	    requested_at,
	    requested_by_user_id
	FROM apps_github_required_user_action
	WHERE
		team_id = $1 AND
		action_user_id = $2;
`,
		teamID, actionUserID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var internalErr *errs.Error
	requiredActions := make([]entity.GithubRequiredUserAction, 0)
	for rows.Next() {
		requiredAction := entity.GithubRequiredUserAction{}
		err = rows.Scan(
			&requiredAction.TeamID,
			&requiredAction.ActionUserID,
			&requiredAction.UserActionType,
			&requiredAction.IsCompleted,
			&requiredAction.RequestedAt,
			&requiredAction.RequestedByUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		requiredActions = append(requiredActions, requiredAction)
	}

	return requiredActions, internalErr
}

func (g GithubRequiredUserAction) FindRequiredUserActionByActionTypeAndUserID(
	ct context.Context,
	teamID uint64,
	actionType entity.GithubUserActionType,
	actionUserID uint64,
) (entity.GithubRequiredUserAction, *errs.Error) {
	requiredAction := entity.GithubRequiredUserAction{}
	err := g.db.QueryRow(`
	SELECT
	    team_id,
	    action_user_id,
	    user_action_type,
	    is_completed,
	    requested_at,
	    requested_by_user_id
	FROM apps_github_required_user_action
	WHERE
		team_id = $1 AND
	    user_action_type = $2 AND
	    action_user_id = $3;
	`,
		teamID,
		actionType,
		actionUserID).
		Scan(
			&requiredAction.TeamID,
			&requiredAction.ActionUserID,
			&requiredAction.UserActionType,
			&requiredAction.IsCompleted,
			&requiredAction.RequestedAt,
			&requiredAction.RequestedByUserID,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.GithubRequiredUserAction{}, errs.NewError(errs.NotFound, err.Error())
		}

		return entity.GithubRequiredUserAction{}, errs.NewError(errs.Unknown, err.Error())
	}

	return requiredAction, nil
}

func (g GithubRequiredUserAction) CreateRequiredUserAction(
	ct context.Context,
	requiredUserAction entity.GithubRequiredUserAction,
) *errs.Error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_required_user_action
	(
	    team_id,
	    action_user_id,
	    user_action_type,
	    is_completed,
	    requested_at,
	    requested_by_user_id
	)
	VALUES ($1, $2, $3, $4, $5, $6);
`,
		requiredUserAction.TeamID,
		requiredUserAction.ActionUserID,
		requiredUserAction.UserActionType,
		requiredUserAction.IsCompleted,
		requiredUserAction.RequestedAt,
		requiredUserAction.RequestedByUserID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g GithubRequiredUserAction) UpdateRequiredUserAction(
	ct context.Context,
	requiredUserAction entity.GithubRequiredUserAction,
) *errs.Error {
	_, err := g.db.Exec(`
		UPDATE apps_github_required_user_action
		SET
		    team_id = $1,
		    action_user_id = $2,
		    user_action_type = $3,
		    is_completed = $4,
		    requested_at = $5,
		    requested_by_user_id = $6
		WHERE
			team_id = $7 AND
			user_action_type=$8 AND
			action_user_id=$9;`,
		requiredUserAction.TeamID,
		requiredUserAction.ActionUserID,
		requiredUserAction.UserActionType,
		requiredUserAction.IsCompleted,
		requiredUserAction.RequestedAt,
		requiredUserAction.RequestedByUserID,
		requiredUserAction.TeamID,
		requiredUserAction.UserActionType,
		requiredUserAction.ActionUserID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g GithubRequiredUserAction) DeleteRequiredUserActionsByActionTypeAndUserID(
	ct context.Context,
	actionType entity.GithubUserActionType,
	actionUserID uint64,
) *errs.Error {
	_, err := g.db.Exec(`
	DELETE FROM apps_github_required_user_action
	WHERE
		user_action_type = $1 AND
		action_user_id = $2;
`,
		actionType,
		actionUserID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewGithubRequiredUserAction(logger telemetry.Logger, sqlDB *sql.DB) GithubRequiredUserAction {
	return GithubRequiredUserAction{logger: logger, db: sqlDB}
}
