package sqldb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type SignInSession struct {
	db *sql.DB
}

var _ dao.SignInSession = (*SignInSession)(nil)

func (s SignInSession) FindSignInSessionByID(sessionID uint64) (entity.SignInSession, error) {
	row := s.db.QueryRow(`
	SELECT 
	    id,
	    redirect_url,
	    type,
	    internal_user_id
	FROM identity_sign_in_session
	WHERE id = $1;
	`,
		sessionID)

	var signInSession entity.SignInSession
	err := row.Scan(
		&signInSession.ID,
		&signInSession.RedirectURL,
		&signInSession.Type,
		&signInSession.InternalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.SignInSession{}, dao.ErrNotFound(fmt.Sprintf(
			"sign in session not found: sessionID=%v",
			sessionID))
	}

	return signInSession, err
}

func (s SignInSession) CreateSignInSession(session entity.SignInSession) error {
	_, err := s.db.Exec(`
	INSERT INTO identity_sign_in_session 
	(
	 	id,
	 	redirect_url,
	 	type,
	 	internal_user_id
	)
	VALUES ($1, $2, $3, $4);
	`,
		session.ID,
		session.RedirectURL,
		session.Type,
		session.InternalUserID)
	return err
}

func (s SignInSession) UpdateSignInSession(session entity.SignInSession) error {
	_, err := s.db.Exec(`
	UPDATE identity_sign_in_session
	SET 
	    redirect_url = $1,
	    type = $2,
	    internal_user_id = $3
	WHERE id = $4;
	`,
		session.RedirectURL,
		session.Type,
		session.InternalUserID,
		session.ID)
	return err
}

func (s SignInSession) DeleteSignInSession(sessionID uint64) error {
	_, err := s.db.Exec(`
	DELETE 
	FROM identity_sign_in_session
	WHERE id = $1;`,
		sessionID)
	return err
}

func NewSignInSession(sqlDB *sql.DB) SignInSession {
	return SignInSession{
		db: sqlDB,
	}
}
