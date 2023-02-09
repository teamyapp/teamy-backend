package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type User struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.User = (*User)(nil)

func (u User) FindUserByID(ct context.Context, userID uint64) (entity.User, *errs.Error) {
	statement := `
	SELECT
		id,
		first_name,
		last_name,
		profile_url,
		created_at,
		updated_at
	FROM "user"
	WHERE id = $1;
`
	user := entity.User{}
	err := u.db.QueryRow(statement, userID).
		Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.ProfileURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("user not found: userID=%v", userID),
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.User{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.User{}, internalErr
	}

	return user, nil
}

func (u User) FindUsersByIDs(ct context.Context, userIDs []uint64) ([]entity.User, *errs.Error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	idsString := toIDsString(userIDs)
	query := fmt.Sprintf(`
	SELECT
		id,
		first_name,
		last_name,
		profile_url,
		created_at,
		updated_at
	FROM "user"
	WHERE id IN (%v)`, idsString)
	rows, err := u.db.Query(query)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	var users []entity.User
	for rows.Next() {
		var user entity.User
		err = rows.
			Scan(
				&user.ID,
				&user.FirstName,
				&user.LastName,
				&user.ProfileURL,
				&user.CreatedAt,
				&user.UpdatedAt,
			)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: newInternalErr})
			continue
		}

		users = append(users, user)
	}

	return users, internalErr
}

func (u User) CreateUser(ct context.Context, user entity.User) *errs.Error {
	_, err := u.db.Exec(`
		INSERT INTO "user"
		(
			id,
			first_name,
			last_name,
			profile_url,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5);`,
		user.ID,
		user.FirstName,
		user.LastName,
		user.ProfileURL,
		user.CreatedAt,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (u User) UpdateUser(ct context.Context, user entity.User) *errs.Error {
	_, err := u.db.Exec(`
		UPDATE "user"
		SET
			first_name = $1,
			last_name = $2,
			profile_url = $3,
			updated_at = $4
		WHERE id = $5;`,
		user.FirstName,
		user.LastName,
		user.ProfileURL,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewUser(dataCollector telemetry.DataCollector, sqlDB *sql.DB) User {
	return User{dataCollector: dataCollector, db: sqlDB}
}
