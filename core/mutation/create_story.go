package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateStory struct {
	storyID          uint64
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	storyDao         dao.Story
	story            entity.Story
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateStory)(nil)

func (c *CreateStory) GetID() uint64 {
	return c.storyID
}

func (c *CreateStory) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	internalErr := c.storyDao.CreateStory(ct, tx, c.story)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func (c *CreateStory) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifierPrepared {
		return nil
	}

	var internalErr *errs.Error
	c.clientNotifiers, internalErr = c.stateSyncer.GetClientNotifiersByTeamID(ct, c.story.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	c.notifierPrepared = true
	return nil
}

func (c *CreateStory) Undo() *errs.Error {
	return nil
}

func (c *CreateStory) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateStory) Clenup(ctx context.Context) *errs.Error {
	return nil
}

func (c *CreateStory) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.storyID,
		CollectionType: realtime.StoryCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.story,
	}
}

func NewCreateStory(
	storyID uint64,
	logger telemetry.Logger,
	stateSyncer *realtime.StateSyncer,
	storyDao dao.Story,
	story entity.Story,
) *CreateStory {
	return &CreateStory{
		storyID:     storyID,
		logger:      logger,
		stateSyncer: stateSyncer,
		storyDao:    storyDao,
		story:       story,
	}
}
