package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type Resource struct {
	db *sql.DB
}

var _ dao.Resource = (*Resource)(nil)

func (r Resource) FindResource(resourceTypeName string, resourceID uint64) (entity.Resource, error) {
	resource := entity.Resource{}
	err := r.db.QueryRow(`
		SELECT
			resource_type,
			resource_id,
			created_at,
			creator_user_id
		FROM resource
		WHERE resource_type = $1 AND resource_id = $2;`,
		resourceTypeName, resourceID).
		Scan(
			&resource.ResourceTypeName,
			&resource.ResourceID,
			&resource.CreatedAt,
			&resource.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Resource{}, dao.ErrNotFound(fmt.Sprintf(
			"resource not found: resource_type=%v, resource_id=%d",
			resourceTypeName, resourceID))
	}

	return resource, err
}

func (r Resource) FindAllResources() ([]entity.Resource, error) {
	rows, err := r.db.Query(`
	SELECT
		resource_type,
		resource_id,
		created_at,
		creator_user_id
	FROM resource;
`)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	resources := make([]entity.Resource, 0)
	for rows.Next() {
		resource := entity.Resource{}
		err = rows.Scan(
			&resource.ResourceTypeName,
			&resource.ResourceID,
			&resource.CreatedAt,
			&resource.CreatorUserID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		resources = append(resources, resource)
	}

	return resources, err
}

func (r Resource) CreateResource(resource entity.Resource) error {
	_, err := r.db.Exec(`
		INSERT INTO resource
		(
			resource_type,
		 	resource_id,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4);`,
		resource.ResourceTypeName,
		resource.ResourceID,
		resource.CreatedAt,
		resource.CreatorUserID,
	)
	return err
}

func (r Resource) DeleteResource(resourceTypeName string, resourceID uint64) error {
	_, err := r.db.Exec(`
		DELETE FROM resource
		WHERE resource_type = $1 AND resource_id = $2;
		`,
		resourceTypeName, resourceID)
	return err
}

func NewResource(sqlDB *sql.DB) Resource {
	return Resource{db: sqlDB}
}
