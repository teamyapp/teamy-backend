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

func (s SignInSession) FindByID(sessionID uint64) (entity.SignInSession, error) {
	row := s.db.QueryRow(`
SELECT id, redirect_url
FROM identity_sign_in_session
WHERE id = $1;
`,
		sessionID)

	var signInSession entity.SignInSession
	err := row.Scan(&signInSession.ID, &signInSession.RedirectURL)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.SignInSession{}, dao.ErrNotFound(fmt.Sprintf(
			"sign in session not found: sessionID=%v",
			sessionID))
	}

	return signInSession, err
}

func (s SignInSession) Add(session entity.SignInSession) error {
	_, err := s.db.Exec(`
INSERT INTO identity_sign_in_session (id, redirect_url)
VALUES ($1, $2);
`,
		session.ID,
		session.RedirectURL)
	return err
}

func (s SignInSession) Update(session entity.SignInSession) error {
	_, err := s.db.Exec(`
UPDATE identity_sign_in_session
SET redirect_url = $1
WHERE id = $2;
`,
		session.RedirectURL,
		session.ID)
	return err
}

func NewSignInSession(sqlDB *sql.DB) SignInSession {
	return SignInSession{
		db: sqlDB,
	}
}
