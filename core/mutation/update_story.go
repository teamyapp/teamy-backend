package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateStory struct {
	logger                  telemetry.Logger
	stateSyncer             *realtime.StateSyncer
	storyDao                dao.Story
	projectDao              dao.Project
	projectStoryRelationDao dao.ProjectStoryRelation
	id                      uint64
	story                   entity.Story
	clientNotifiers         []*realtime.ClientNotifier
	notifierPrepared        bool
}

var _ realtime.Mutation = (*UpdateStory)(nil)

func (u *UpdateStory) GetID() uint64 {
	return u.id
}

func (u *UpdateStory) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return u.storyDao.UpdateStory(ct, tx, u.story)
}

func (u *UpdateStory) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if u.notifierPrepared {
		return nil
	}

	projectIDs, err := u.projectStoryRelationDao.FindProjectIDsByStoryIDWithTx(ct, tx, u.story.ID)
	if err != nil {
		return err
	}

	projects, err := u.projectDao.FindProjectsByIDsWithTx(ct, tx, projectIDs)
	if err != nil {
		return err
	}

	teamIDs := collect.Map(projects, func(p entity.Project, _ int) uint64 {
		return p.TeamID
	})
	u.clientNotifiers, err = u.stateSyncer.GetClientNotifiersByTeamIDs(ct, teamIDs)
	if err != nil {
		return err
	}

	u.notifierPrepared = true
	return nil
}

func (u *UpdateStory) Undo() *errs.Error {
	return nil
}

func (u *UpdateStory) GetClientNotifiers() []*realtime.ClientNotifier {
	return u.clientNotifiers
}

func (u *UpdateStory) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.StoryCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.story,
	}
}

func (u *UpdateStory) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateStory(
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	storyDao dao.Story,
	projectDao dao.Project,
	projectStoryRelationDao dao.ProjectStoryRelation,
	story entity.Story,
) *UpdateStory {
	return &UpdateStory{
		logger:                  logger,
		stateSyncer:             stateSyncer,
		storyDao:                storyDao,
		projectDao:              projectDao,
		projectStoryRelationDao: projectStoryRelationDao,
		id:                      stateSyncer.NextMutationID(),
		story:                   story,
	}
}
