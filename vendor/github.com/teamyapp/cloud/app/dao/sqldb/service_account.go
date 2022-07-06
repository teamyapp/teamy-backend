package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type ServiceAccount struct {
	db *sql.DB
}

var _ dao.ServiceAccount = (*ServiceAccount)(nil)

func (s ServiceAccount) FindAllServiceAccounts(accountOwnerID uint64) ([]entity.ServiceAccount, error) {
	rows, err := s.db.Query(`
	SELECT
	    id,
	    name,
	    owner_user_id,
	    secret,
	    created_at
	FROM identity_service_account
	WHERE owner_user_id = $1;`,
		accountOwnerID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	serviceAccounts := make([]entity.ServiceAccount, 0)
	for rows.Next() {
		serviceAccount := entity.ServiceAccount{}
		err = rows.Scan(
			&serviceAccount.ID,
			&serviceAccount.Name,
			&serviceAccount.OwnerUserID,
			&serviceAccount.Secret,
			&serviceAccount.CreatedAt,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		serviceAccounts = append(serviceAccounts, serviceAccount)
	}

	return serviceAccounts, err
}

func (s ServiceAccount) FindServiceAccountByID(serviceAccountID uint64) (entity.ServiceAccount, error) {
	serviceAccount := entity.ServiceAccount{}
	err := s.db.QueryRow(`
	SELECT
	    id,
	    name,
	    owner_user_id,
	    secret,
	    created_at
	FROM identity_service_account
	WHERE id = $1;`,
		serviceAccountID).
		Scan(
			&serviceAccount.ID,
			&serviceAccount.Name,
			&serviceAccount.OwnerUserID,
			&serviceAccount.Secret,
			&serviceAccount.CreatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.ServiceAccount{}, dao.ErrNotFound(fmt.Sprintf(
			"service account not found: id=%v", serviceAccountID))
	}

	return serviceAccount, err
}

func (s ServiceAccount) CreateServiceAccount(serviceAccount entity.ServiceAccount) error {
	_, err := s.db.Exec(`
	INSERT INTO identity_service_account
	(
	 	id,
	 	name,
	 	owner_user_id,
	 	secret,
	 	created_at
	)
	VALUES ($1, $2, $3, $4, $5);`,
		int64(serviceAccount.ID),
		serviceAccount.Name,
		serviceAccount.OwnerUserID,
		serviceAccount.Secret,
		serviceAccount.CreatedAt,
	)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (s ServiceAccount) UpdateServiceAccount(serviceAccount entity.ServiceAccount) error {
	_, err := s.db.Exec(`
	UPDATE identity_service_account
	SET
	    id = $1,
	    name = $2,
	    owner_user_id = $3,
	    secret = $4,
	    created_at = $5
	WHERE id = $6;`,
		serviceAccount.ID,
		serviceAccount.Name,
		serviceAccount.OwnerUserID,
		serviceAccount.Secret,
		serviceAccount.CreatedAt,
		serviceAccount.ID,
	)
	return err
}

func (s ServiceAccount) DeleteServiceAccount(serviceAccountID uint64) error {
	_, err := s.db.Exec(`
		DELETE 
		FROM identity_service_account
		WHERE id = $1;`,
		serviceAccountID)
	return err
}

func NewServiceAccount(sqlDB *sql.DB) ServiceAccount {
	return ServiceAccount{db: sqlDB}
}
