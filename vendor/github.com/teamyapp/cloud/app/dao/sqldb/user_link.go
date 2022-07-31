package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type UserLink struct {
	db *sql.DB
}

var _ dao.UserLink = (*UserLink)(nil)

func (u UserLink) FindUserLinkByExternalUserID(authProvider string, externalUserID string) (entity.UserLink, error) {
	row := u.db.QueryRow(`
		SELECT
		    auth_provider,
		    external_user_id,
		    external_user_label,
		    internal_user_id
		FROM identity_user_link
		WHERE auth_provider = $1 AND external_user_id = $2;
`,
		authProvider,
		externalUserID)

	var userLink entity.UserLink
	err := row.Scan(
		&userLink.AuthProvider,
		&userLink.ExternalUserID,
		&userLink.ExternalUserLabel,
		&userLink.InternalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.UserLink{}, dao.ErrNotFound(fmt.Sprintf(
			"user link not found: authProvider=%v externalUserID=%v",
			authProvider,
			externalUserID))
	}

	return userLink, err
}

func (u UserLink) FindUserLinksByInternalUserID(internalUserID uint64) ([]entity.UserLink, error) {
	rows, err := u.db.Query(
		`
		SELECT
		    auth_provider,
		    external_user_id,
		    external_user_label,
		    internal_user_id
		FROM identity_user_link
		WHERE internal_user_id = $1;
`,
		internalUserID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	userLinks := make([]entity.UserLink, 0)
	for rows.Next() {
		userLink := entity.UserLink{}
		err = rows.Scan(
			&userLink.AuthProvider,
			&userLink.ExternalUserID,
			&userLink.ExternalUserLabel,
			&userLink.InternalUserID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		userLinks = append(userLinks, userLink)
	}

	return userLinks, err
}

func (u UserLink) CreateUserLink(userLink entity.UserLink) error {
	_, err := u.db.Exec(`
		INSERT INTO identity_user_link 
		(
		 	auth_provider,
		 	external_user_id,
		 	external_user_label,
		 	internal_user_id
		)
		VALUES ($1, $2, $3, $4);
		`,
		userLink.AuthProvider,
		userLink.ExternalUserID,
		userLink.ExternalUserLabel,
		userLink.InternalUserID)
	return err
}

func (u UserLink) DeleteUserLink(authProvider string, internalUserID uint64) error {
	_, err := u.db.Exec(`
		DELETE 
		FROM identity_user_link
		WHERE auth_provider = $1 AND internal_user_id = $2;`,
		authProvider,
		internalUserID)
	return err
}

func NewUserLink(sqlDB *sql.DB) UserLink {
	return UserLink{
		db: sqlDB,
	}
}
