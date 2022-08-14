package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type UserGroup struct {
	db *sql.DB
}

var _ dao.UserGroup = (*UserGroup)(nil)

func (u UserGroup) FindGroupByID(groupID uint64) (entity.UserGroup, error) {
	group := entity.UserGroup{}
	err := u.db.QueryRow(`
		SELECT
			id,
			name,
			description,
			created_at,
			creator_user_id,
			updated_at
		FROM user_group
		WHERE id = $1;`,
		groupID).
		Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.CreatedAt,
			&group.CreatorUserID,
			&group.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.UserGroup{}, dao.ErrNotFound(fmt.Sprintf(
			"user group not found: group_id=%d",
			groupID))
	}

	return group, err
}

func (u UserGroup) FindAllGroups() ([]entity.UserGroup, error) {
	rows, err := u.db.Query(`
		SELECT
			id,
			name,
			description,
			created_at,
			creator_user_id,
			updated_at
		FROM user_group;
	`)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	groups := make([]entity.UserGroup, 0)
	for rows.Next() {
		group := entity.UserGroup{}
		err = rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.CreatedAt,
			&group.CreatorUserID,
			&group.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		groups = append(groups, group)
	}

	return groups, err
}

func (u UserGroup) CreateGroup(group entity.UserGroup) error {
	_, err := u.db.Exec(`
		INSERT INTO user_group
		(
			id,
			name,
			description,
			created_at,
			creator_user_id,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		group.ID,
		group.Name,
		group.Description,
		group.CreatedAt,
		group.CreatorUserID,
		group.UpdatedAt,
	)
	return err
}

func (u UserGroup) UpdateGroup(group entity.UserGroup) error {
	_, err := u.db.Exec(`
		UPDATE user_group
		SET
			name = $1,
			description = $2,
			created_at = $3,
			creator_user_id = $4,
			updated_at = $5
		WHERE id = $6;`,
		group.Name,
		group.Description,
		group.CreatedAt,
		group.CreatorUserID,
		group.UpdatedAt,
		group.ID,
	)
	return err
}

func (u UserGroup) DeleteGroup(groupID uint64) error {
	_, err := u.db.Exec(`
		DELETE FROM user_group
		WHERE id = $1;
		`,
		groupID)
	return err
}

func NewUserGroup(sqlDB *sql.DB) UserGroup {
	return UserGroup{db: sqlDB}
}
