package sqldb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type UserLink struct {
	db *sql.DB
}

var _ dao.UserLink = (*UserLink)(nil)

func (u UserLink) FindByExternalUserID(authProvider string, externalUserID string) (entity.UserLink, error) {
	row := u.db.QueryRow(`
SELECT auth_provider, external_user_id, internal_user_id 
FROM identity_user_link
WHERE auth_provider = $1 AND external_user_id = $2;
`,
		authProvider,
		externalUserID)

	var userLink entity.UserLink
	err := row.Scan(&userLink.AuthProvider, &userLink.ExternalUserID, &userLink.InternalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.UserLink{}, dao.ErrNotFound(fmt.Sprintf(
			"user link not found: authProvider=%v externalUserID=%v",
			authProvider,
			externalUserID))
	}

	return userLink, err
}

func (u UserLink) Add(userLink entity.UserLink) error {
	_, err := u.db.Exec(`
INSERT INTO identity_user_link (auth_provider, external_user_id, internal_user_id)
VALUES ($1, $2, $3);
`,
		userLink.AuthProvider,
		userLink.ExternalUserID,
		userLink.InternalUserID)
	return err
}

func NewUserLink(sqlDB *sql.DB) UserLink {
	return UserLink{
		db: sqlDB,
	}
}
