package daotestv2

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintParticipant struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

func (s SprintParticipant) FindParticipantIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) FindParticipantsBySprintID(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) FindParticipant(ct context.Context, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) FindParticipantIDsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error) {
	var participantIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintParticipantTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currSprintParticipant := rawRow.(entity.SprintParticipant)
				if currSprintParticipant.SprintID == sprintID {
					participantIDs = append(participantIDs, currSprintParticipant.UserID)
				}
			}

			return nil
		},
	})
	return participantIDs, err
}

func (s SprintParticipant) FindParticipantsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (s SprintParticipant) FindParticipantWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64, participantUserID uint64) (entity.SprintParticipant, *errs.Error) {
	var participant entity.SprintParticipant
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintParticipantTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currSprintParticipant := rawRow.(entity.SprintParticipant)
				if currSprintParticipant.SprintID == sprintID && currSprintParticipant.UserID == participantUserID {
					participant = currSprintParticipant
					return nil
				}
			}

			return errs.NewError(errs.NotFound, fmt.Sprintf("Participant not found: sprintId=%v participantUserID=%v", sprintID, participantUserID))
		},
	})

	if err != nil {
		return entity.SprintParticipant{}, err
	}

	return participant, nil
}

func (s SprintParticipant) CreateSprintParticipant(ct context.Context, tx *transaction.Transaction, participant entity.SprintParticipant) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintParticipantTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currSprintParticipant := rawRow.(entity.SprintParticipant)
				if currSprintParticipant.SprintID == participant.SprintID && currSprintParticipant.UserID == participant.UserID {
					return errs.NewError(errs.AlreadyExists, fmt.Sprintf("Participant already exists: sprintId=%v participantUserID=%v", participant.SprintID, participant.UserID))
				}
			}

			table.Rows = append(table.Rows, participant)
			return nil
		}, Undo: func() *errs.Error {
			table, err := s.db.GetTable(SprintParticipantTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currSprintParticipant := rawRow.(entity.SprintParticipant)
				if currSprintParticipant.SprintID == participant.SprintID && currSprintParticipant.UserID == participant.UserID {
					continue
				}
				rows = append(rows, currSprintParticipant)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (s SprintParticipant) UpdateSprintParticipant(ct context.Context, tx *transaction.Transaction, participant entity.SprintParticipant) *errs.Error {
	oldParticipant, err := s.FindParticipantWithTx(ct, tx, participant.SprintID, participant.UserID)
	if err != nil {
		return err
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintParticipantTableName)
			if err != nil {
				return err
			}

			for i, rawRow := range table.Rows {
				currSprintParticipant := rawRow.(entity.SprintParticipant)
				if currSprintParticipant.SprintID == participant.SprintID && currSprintParticipant.UserID == participant.UserID {
					table.Rows[i] = participant
					return nil
				}
			}

			return errs.NewError(errs.NotFound, fmt.Sprintf("Participant not found: sprintId=%v participantUserID=%v", participant.SprintID, participant.UserID))
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(SprintParticipantTableName)
			if err != nil {
				return err
			}

			for i, rawRow := range table.Rows {
				currSprintParticipant := rawRow.(entity.SprintParticipant)
				if currSprintParticipant.SprintID == participant.SprintID && currSprintParticipant.UserID == participant.UserID {
					table.Rows[i] = oldParticipant
					return nil
				}
			}

			return nil
		},
	})
}

func (s SprintParticipant) DeleteSprintParticipant(ct context.Context, tx *transaction.Transaction, sprintID uint64, userID uint64) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintParticipantTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currSprintParticipant := rawRow.(entity.SprintParticipant)
				if currSprintParticipant.SprintID == sprintID && currSprintParticipant.UserID == userID {
					continue
				}
				rows = append(rows, currSprintParticipant)
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(SprintParticipantTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, entity.SprintParticipant{
				SprintID: sprintID,
				UserID:   userID,
			})

			return nil
		},
	})
}

func NewSprintParticipant(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) SprintParticipant {
	return SprintParticipant{db: db, transactionFactory: transactionFactory}
}
