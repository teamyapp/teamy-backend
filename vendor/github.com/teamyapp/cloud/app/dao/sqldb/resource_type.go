package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type ResourceType struct {
	db *sql.DB
}

var _ dao.ResourceType = (*ResourceType)(nil)

func (r ResourceType) FindResourceType(resourceTypeName string) (entity.ResourceType, error) {
	resourceTypeEntity := entity.ResourceType{}
	err := r.db.QueryRow(`
		SELECT
			resource_type,
			created_at,
			creator_user_id
		FROM resource_type
		WHERE resource_type = $1;`,
		resourceTypeName).
		Scan(
			&resourceTypeEntity.ResourceTypeName,
			&resourceTypeEntity.CreatedAt,
			&resourceTypeEntity.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.ResourceType{}, dao.ErrNotFound(fmt.Sprintf(
			"resource type not found: resource_type=%v",
			resourceTypeName))
	}

	return resourceTypeEntity, err
}

func (r ResourceType) FindAllResourceTypes() ([]entity.ResourceType, error) {
	rows, err := r.db.Query(`
	SELECT
		resource_type,
		created_at,
		creator_user_id
	FROM resource_type;
`)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	resourceTypeEntities := make([]entity.ResourceType, 0)
	for rows.Next() {
		resourceTypeEntity := entity.ResourceType{}
		err = rows.Scan(
			&resourceTypeEntity.ResourceTypeName,
			&resourceTypeEntity.CreatedAt,
			&resourceTypeEntity.CreatorUserID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		resourceTypeEntities = append(resourceTypeEntities, resourceTypeEntity)
	}

	return resourceTypeEntities, err
}

func (r ResourceType) CreateResourceType(resourceTypeEntity entity.ResourceType) error {
	_, err := r.db.Exec(`
		INSERT INTO resource_type
		(
			resource_type,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3);`,
		resourceTypeEntity.ResourceTypeName,
		resourceTypeEntity.CreatedAt,
		resourceTypeEntity.CreatorUserID,
	)
	return err
}

func (r ResourceType) DeleteResourceType(resourceTypeName string) error {
	_, err := r.db.Exec(`
		DELETE FROM resource_type
		WHERE resource_type = $1;
		`,
		resourceTypeName)
	return err
}

func NewResourceType(sqlDB *sql.DB) ResourceType {
	return ResourceType{db: sqlDB}
}
