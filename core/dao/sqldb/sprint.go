package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.Sprint = (*Sprint)(nil)

func (s Sprint) FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, error) {
	sprint := entity.Sprint{}
	err := s.db.QueryRow(`
		SELECT
			id,
			start_at,
			end_at,
			created_at,
			owning_team_id
		FROM sprint
		WHERE id = $1;`,
		sprintID).
		Scan(
			&sprint.ID,
			&sprint.StartAt,
			&sprint.EndAt,
			&sprint.CreatedAt,
			&sprint.OwningTeamID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.Sprint{}, dao.ErrNotFound(fmt.Sprintf(
			"sprint not found: id=%v",
			sprintID))
	}

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return sprint, err
}

func (s Sprint) FindSprintsByIDs(ct context.Context, sprintIDs []uint64) ([]entity.Sprint, error) {
	if len(sprintIDs) == 0 {
		return []entity.Sprint{}, nil
	}

	idsString := toIDsString(sprintIDs)
	query := fmt.Sprintf(`
	SELECT
		id,
		start_at,
		end_at,
		created_at,
		owning_team_id
	FROM sprint
	WHERE id IN (%s);`, idsString)
	rows, err := s.db.Query(query)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	var sprints []entity.Sprint
	for rows.Next() {
		var sprint entity.Sprint
		err = rows.
			Scan(
				&sprint.ID,
				&sprint.StartAt,
				&sprint.EndAt,
				&sprint.CreatedAt,
				&sprint.OwningTeamID,
			)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
}

func (s Sprint) FindSprintsByTeamID(ct context.Context, teamID uint64) ([]entity.Sprint, error) {
	rows, err := s.db.Query(
		`
	SELECT
		id,
		start_at,
		end_at,
		created_at,
		owning_team_id
	FROM sprint
	WHERE owning_team_id = $1;
`,
		teamID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	var sprints []entity.Sprint
	for rows.Next() {
		var sprint entity.Sprint
		err = rows.
			Scan(
				&sprint.ID,
				&sprint.StartAt,
				&sprint.EndAt,
				&sprint.CreatedAt,
				&sprint.OwningTeamID,
			)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
}

func (s Sprint) FindAllSprints(ct context.Context) ([]entity.Sprint, error) {
	rows, err := s.db.Query(`
	SELECT
		id,
		start_at,
		end_at,
		created_at,
		owning_team_id
	FROM sprint;
`)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	var sprints []entity.Sprint
	for rows.Next() {
		var sprint entity.Sprint
		err = rows.
			Scan(
				&sprint.ID,
				&sprint.StartAt,
				&sprint.EndAt,
				&sprint.CreatedAt,
				&sprint.OwningTeamID,
			)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		sprints = append(sprints, sprint)
	}

	return sprints, nil
}

func (s Sprint) CreateSprint(ct context.Context, sprint entity.Sprint) error {
	_, err := s.db.Exec(`
		INSERT INTO sprint
		(
			id,
			start_at,
			end_at,
			created_at,
			owning_team_id
		)
		VALUES ($1, $2, $3, $4, $5);`,
		sprint.ID,
		sprint.StartAt,
		sprint.EndAt,
		sprint.CreatedAt,
		sprint.OwningTeamID,
	)

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (s Sprint) DeleteSprint(ct context.Context, sprintID uint64) error {
	_, err := s.db.Exec(`
		DELETE FROM sprint
		WHERE id = $1;
		`,
		sprintID)

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func NewSprint(dataCollector telemetry.DataCollector, sqlDB *sql.DB) Sprint {
	return Sprint{dataCollector: dataCollector, db: sqlDB}
}
