package sqldb

import (
	"database/sql"
	"log"

	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubRequiredUserAction struct {
	db *sql.DB
}

var _ dao.GithubRequiredUserAction = (*GithubRequiredUserAction)(nil)

func (g GithubRequiredUserAction) FindRequiredUserActionsByActionUserID(
	teamID uint64,
	actionUserID uint64,
) ([]entity.GithubRequiredUserAction, error) {
	rows, err := g.db.Query(`
	SELECT
		id,
	    team_id,
	    action_user_id,
	    user_action_type,
	    is_completed,
	    requested_at,
	    requested_by_user_id
	FROM apps_github_required_user_action
	WHERE team_id = $1 AND action_user_id = $2;
`,
		teamID, actionUserID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	requiredActions := make([]entity.GithubRequiredUserAction, 0)
	for rows.Next() {
		requiredAction := entity.GithubRequiredUserAction{}
		err = rows.Scan(
			&requiredAction.ID,
			&requiredAction.TeamID,
			&requiredAction.ActionUserID,
			&requiredAction.UserActionType,
			&requiredAction.IsCompleted,
			&requiredAction.RequestedAt,
			&requiredAction.RequestedByUserID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		requiredActions = append(requiredActions, requiredAction)
	}

	return requiredActions, err
}

func (g GithubRequiredUserAction) CreateRequiredUserAction(requiredUserAction entity.GithubRequiredUserAction) error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_required_user_action
	(
	    id,
	    team_id,
	    action_user_id,
	    user_action_type,
	    is_completed,
	    requested_at,
	    requested_by_user_id
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7);
`,
		requiredUserAction.ID,
		requiredUserAction.TeamID,
		requiredUserAction.ActionUserID,
		requiredUserAction.UserActionType,
		requiredUserAction.IsCompleted,
		requiredUserAction.RequestedAt,
		requiredUserAction.RequestedByUserID,
	)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (g GithubRequiredUserAction) UpdateRequiredUserAction(requiredUserAction entity.GithubRequiredUserAction) error {
	_, err := g.db.Exec(`
		UPDATE apps_github_required_user_action
		SET
		    team_id = $1,
		    action_user_id = $2,
		    user_action_type = $3,
		    is_completed = $4,
		    requested_at = $5,
		    requested_by_user_id = $6
		WHERE id = $7;`,
		requiredUserAction.TeamID,
		requiredUserAction.ActionUserID,
		requiredUserAction.UserActionType,
		requiredUserAction.IsCompleted,
		requiredUserAction.RequestedAt,
		requiredUserAction.RequestedByUserID,
		requiredUserAction.ID,
	)
	return err
}

func NewGithubRequiredUserAction(sqlDB *sql.DB) GithubRequiredUserAction {
	return GithubRequiredUserAction{db: sqlDB}
}
