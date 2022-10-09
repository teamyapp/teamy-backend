package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.SprintParticipant = (*SprintParticipant)(nil)

func (s SprintParticipant) FindParticipantIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, error) {
	rows, err := s.db.Query(
		`
	SELECT
		user_id
	FROM sprint_participant
	WHERE sprint_id = $1;
`,
		sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	participantUserIDs := make([]uint64, 0)
	for rows.Next() {
		var participantUserID uint64
		err = rows.Scan(
			&participantUserID,
		)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		participantUserIDs = append(participantUserIDs, participantUserID)
	}

	return participantUserIDs, nil
}

func (s SprintParticipant) FindParticipantsBySprintID(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, error) {
	rows, err := s.db.Query(
		`
	SELECT
		sprint_id,
		user_id,
		total_bandwidth,
	 	unused_bandwidth,
		created_at,
	 	updated_at
	FROM sprint_participant
	WHERE sprint_id = $1;
`,
		sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	defer rows.Close()
	sprintParticipants := make([]entity.SprintParticipant, 0)
	for rows.Next() {
		sprintParticipant := entity.SprintParticipant{}
		err = rows.Scan(
			&sprintParticipant.SprintID,
			&sprintParticipant.UserID,
			&sprintParticipant.TotalBandwidth,
			&sprintParticipant.UnusedBandwidth,
			&sprintParticipant.CreatedAt,
			&sprintParticipant.UpdatedAt,
		)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		sprintParticipants = append(sprintParticipants, sprintParticipant)
	}

	return sprintParticipants, nil
}

func (s SprintParticipant) FindParticipant(ct context.Context, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, error) {
	participant := entity.SprintParticipant{}
	err := s.db.QueryRow(`
	SELECT
		sprint_id,
		user_id,
		total_bandwidth,
		unused_bandwidth,
		created_at,
		updated_at
	FROM sprint_participant
	WHERE sprint_id = $1 AND user_id=$2;
`,
		sprintID,
		participantUserID).
		Scan(
			&participant.SprintID,
			&participant.UserID,
			&participant.TotalBandwidth,
			&participant.UnusedBandwidth,
			&participant.CreatedAt,
			&participant.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.SprintParticipant{}, dao.ErrNotFound(fmt.Sprintf(
			"participant not found: sprintID=%v", sprintID))
	}

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return participant, err
}

func (s SprintParticipant) CreateSprintParticipant(ct context.Context, participant entity.SprintParticipant) error {
	_, err := s.db.Exec(`
	INSERT INTO sprint_participant
	(
	    sprint_id,
		user_id,
		total_bandwidth,
	 	unused_bandwidth,
		created_at,
	 	updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6);
`,
		participant.SprintID,
		participant.UserID,
		participant.TotalBandwidth,
		participant.UnusedBandwidth,
		participant.CreatedAt,
		participant.UpdatedAt,
	)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (s SprintParticipant) UpdateSprintParticipant(ct context.Context, participant entity.SprintParticipant) error {
	_, err := s.db.Exec(`
		UPDATE sprint_participant
		SET
		    sprint_id = $1,
			user_id = $2,
			total_bandwidth = $3,
			unused_bandwidth = $4,
			created_at = $5,
			updated_at = $6
		WHERE sprint_id = $7 AND user_id = $8;`,
		participant.SprintID,
		participant.UserID,
		participant.TotalBandwidth,
		participant.UnusedBandwidth,
		participant.CreatedAt,
		participant.UpdatedAt,
		participant.SprintID,
		participant.UserID,
	)

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (s SprintParticipant) DeleteSprintParticipant(ct context.Context, sprintID uint64, userID uint64) error {
	_, err := s.db.Exec(`
		DELETE FROM sprint_participant
		WHERE sprint_id = $1 AND user_id = $2;
		`,
		sprintID,
		userID)

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewSprintParticipant(dataCollector obs.DataCollector, sqlDB *sql.DB) SprintParticipant {
	return SprintParticipant{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
