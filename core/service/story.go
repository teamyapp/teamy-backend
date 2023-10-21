package service

import (
	cloudClient "github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/client"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"time"
)

var awaitStoryStatuses = map[entity.StoryStatus]bool{
	entity.StoryStatusTodo:       true,
	entity.StoryStatusInProgress: true,
	entity.StoryStatusPaused:     true,
	entity.StoryStatusDelivered:  true,
}

//type Story struct {
//	ID               uint64
//	Goal             string
//	DueAt            *time.Time
//	Context          *string
//	CreatorUserID    uint64
//	OwnerUserID      *uint64
//	OwningTeamID     uint64
//	Status           entity.StoryStatus
//	IsPlanned        bool
//	Effort           *time.Duration
//	CommentsThreadID uint64
//	CreatedAt        time.Time
//	UpdatedAt        *time.Time
//	DeliveredAt      *time.Time
//	Priority         *entity.Priority
//}

type createStoryInput struct {
	Goal         string
	DueAt        *time.Time
	Context      *string
	OwnerUserID  *uint64
	OwningTeamID uint64
	IsPlanned    bool
	Effort       *time.Duration
	Priority     *entity.Priority
}

type updateStoryInput struct {
	ID           uint64
	Goal         *string
	DueAt        *time.Time
	Context      *string
	OwnerUserID  *uint64
	OwningTeamID *uint64
	IsPlanned    *bool
	Effort       *time.Duration
	Priority     *entity.Priority
}

type Story struct {
	logger                telemetry.Logger
	cloudClientRegistry   *client.Registry
	authorizer            cloudClient.Authorizer
	featureToggles        feature.Toggles
	transactionFactory    cloudTransaction.Factory
	activityCache         *cache.Activity
	storyDao              dao.Story
	storyLinkDao          dao.StoryLink
	sprintParticipantDao  dao.SprintParticipant
	sprintTaskRelationDao dao.SprintStoryRelation
}
