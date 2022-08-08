package sqldb

import (
	"database/sql"

	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubRequiredUserAction struct {
	db *sql.DB
}

var _ dao.GithubRequiredUserAction = (*GithubRequiredUserAction)(nil)

func (g GithubRequiredUserAction) FindRequiredUserActionsByUserID(
	teamID uint64,
	userID uint64,
) ([]entity.GithubRequiredUserAction, error) {
	//TODO implement me
	panic("implement me")
}

func (g GithubRequiredUserAction) CreateRequiredUserAction(requiredUserAction entity.GithubRequiredUserAction) error {
	//TODO implement me
	panic("implement me")
}

func (g GithubRequiredUserAction) UpdateRequiredUserAction(requiredUserAction entity.GithubRequiredUserAction) error {
	//TODO implement me
	panic("implement me")
}

func NewGithubRequiredUserAction(sqlDB *sql.DB) GithubRequiredUserAction {
	return GithubRequiredUserAction{db: sqlDB}
}
