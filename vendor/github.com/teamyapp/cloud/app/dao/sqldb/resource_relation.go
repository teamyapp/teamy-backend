package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/cloud/app/dao"
	"github.com/teamyapp/cloud/app/entity"
)

type ResourceRelation struct {
	db *sql.DB
}

var _ dao.ResourceRelation = (*ResourceRelation)(nil)

func (r ResourceRelation) FindResourceRelation(
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) (entity.ResourceRelation, error) {
	resourceRelation := entity.ResourceRelation{}
	err := r.db.QueryRow(`
		SELECT
			child_resource_type,
			child_resource_id,
			parent_resource_type,
			parent_resource_id,
			created_at,
			creator_user_id
		FROM resource_relation
		WHERE child_resource_type = $1 AND child_resource_id = $2 AND parent_resource_type = $3 AND parent_resource_id = $4;`,
		childResourceType, childResourceID, parentResourceType, parentResourceID).
		Scan(
			&resourceRelation.ChildResourceType,
			&resourceRelation.ChildResourceID,
			&resourceRelation.ParentResourceType,
			&resourceRelation.ParentResourceID,
			&resourceRelation.CreatedAt,
			&resourceRelation.CreatorUserID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.ResourceRelation{}, dao.ErrNotFound(fmt.Sprintf(
			"resource relation not found: child_resource_type=%v, child_resource_id=%d, parent_resource_type=%v, parent_resource_id=%d",
			childResourceType, childResourceID, parentResourceType, parentResourceID))
	}

	return resourceRelation, err
}

func (r ResourceRelation) FindResourceRelations(childResourceType string, childResourceID uint64) ([]entity.ResourceRelation, error) {
	rows, err := r.db.Query(`
		SELECT
			child_resource_type,
			child_resource_id,
			parent_resource_type,
			parent_resource_id,
			created_at,
			creator_user_id
		FROM resource_relation
		WHERE child_resource_type = $1 AND child_resource_id = $2;`,
		childResourceType, childResourceID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	resourceRelations := make([]entity.ResourceRelation, 0)
	for rows.Next() {
		resourceRelation := entity.ResourceRelation{}
		err = rows.Scan(
			&resourceRelation.ChildResourceType,
			&resourceRelation.ChildResourceID,
			&resourceRelation.ParentResourceType,
			&resourceRelation.ParentResourceID,
			&resourceRelation.CreatedAt,
			&resourceRelation.CreatorUserID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		resourceRelations = append(resourceRelations, resourceRelation)
	}

	return resourceRelations, err
}

func (r ResourceRelation) FindAllResourceRelations() ([]entity.ResourceRelation, error) {
	rows, err := r.db.Query(`
		SELECT
			child_resource_type,
			child_resource_id,
			parent_resource_type,
			parent_resource_id,
			created_at,
			creator_user_id
		FROM resource_relation;
	`)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	resourceRelations := make([]entity.ResourceRelation, 0)
	for rows.Next() {
		resourceRelation := entity.ResourceRelation{}
		err = rows.Scan(
			&resourceRelation.ChildResourceType,
			&resourceRelation.ChildResourceID,
			&resourceRelation.ParentResourceType,
			&resourceRelation.ParentResourceID,
			&resourceRelation.CreatedAt,
			&resourceRelation.CreatorUserID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		resourceRelations = append(resourceRelations, resourceRelation)
	}

	return resourceRelations, err
}

func (r ResourceRelation) CreateResourceRelation(resourceRelation entity.ResourceRelation) error {
	_, err := r.db.Exec(`
		INSERT INTO resource_relation
		(
		 	child_resource_type,
		 	child_resource_id,
		 	parent_resource_type,
		 	parent_resource_id,
			created_at,
			creator_user_id
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		resourceRelation.ChildResourceType,
		resourceRelation.ChildResourceID,
		resourceRelation.ParentResourceType,
		resourceRelation.ParentResourceID,
		resourceRelation.CreatedAt,
		resourceRelation.CreatorUserID,
	)
	return err
}

func (r ResourceRelation) DeleteResourceRelation(
	childResourceType string,
	childResourceID uint64,
	parentResourceType string,
	parentResourceID uint64,
) error {
	_, err := r.db.Exec(`
		DELETE FROM resource_relation
		WHERE child_resource_type = $1 AND child_resource_id = $2 AND parent_resource_type = $3 AND parent_resource_id = $4;
		`,
		childResourceType, childResourceID, parentResourceType, parentResourceID)
	return err
}

func NewResourceRelation(sqlDB *sql.DB) ResourceRelation {
	return ResourceRelation{db: sqlDB}
}
