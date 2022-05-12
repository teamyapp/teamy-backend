package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Mutation struct {
}

/* Task */

func (m Mutation) CreateTask(ct context.Context, args struct {
	TeamID graphql.ID
	Task   struct {
		Goal        string
		Context     *string
		OwnerUserID *graphql.ID
		Status      entity.TaskStatus
		DueAt       *graphql.Time
	}
}) (Task, error) {
	panic("implement me")
}

func (m Mutation) UpdateTask(ct context.Context, args struct {
	TaskID graphql.ID
	Input  struct {
		Goal        *string
		Context     *string
		OwnerUserID *graphql.ID
		Status      *entity.TaskStatus
		DueAt       *graphql.Time
	}
}) (Task, error) {
	panic("implement me")
}

func (m Mutation) DeleteTask(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	panic("implement me")
}

/* Message */

func (m Mutation) CreateMessage(ct context.Context, args struct {
	ThreadID graphql.ID
	Message  struct {
		Body string
	}
}) (Message, error) {
	panic("implement me")
}

func (m Mutation) UpdateMessage(ct context.Context, args struct {
	MessageID graphql.ID
	Message   struct {
		body *string
	}
}) (Message, error) {
	panic("implement me")
}

func (m Mutation) DeleteMessage(ct context.Context, args struct {
	MessageID graphql.ID
}) (Message, error) {
	panic("implement me")
}

/* Team */

func (m Mutation) CreateTeam(ct context.Context, args struct {
	Team struct {
		Name string
	}
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) UpdateTeam(ct context.Context, args struct {
	TeamID graphql.ID
	Input  struct {
		Name    *string
		IconURL *string
	}
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) AddMemberToTeam(ct context.Context, args struct {
	TeamID   graphql.ID
	MemberID graphql.ID
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) RemoveMemberFromTeam(ct context.Context, args struct {
	TeamID   graphql.ID
	MemberID graphql.ID
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) RemoveTaskFromTeam(ct context.Context, args struct {
	TeamID graphql.ID
	TaskID graphql.ID
}) (Team, error) {
	panic("implement me")
}

func (m Mutation) PromoteTeamTaskToNeedAttention(ct context.Context, args struct {
	TeamID graphql.ID
	TaskID graphql.ID
}) (Team, error) {
	panic("implement me")
}

/* User */

func (m Mutation) CreateUser(ct context.Context, args struct {
	User struct {
		LastName   string
		FirstName  *string
		ProfileURL *string
	}
}) (User, error) {
	panic("implement me")
}

func (m Mutation) UpdateUser(ct context.Context, args struct {
	UserID graphql.ID
	Input  struct {
		LastName   *string
		FirstName  *string
		ProfileURL *string
	}
}) (User, error) {
	panic("implement me")
}

/* Invitation */

func (m Mutation) CreateInvitation(ct context.Context, args struct {
	TeamID     graphql.ID
	Invitation struct {
		ReceiverFirstName *string
		ReceiverEmail     *string
		ExpireAT          *graphql.Time
	}
}) (Invitation, error) {
	panic("implement me")
}

func (m Mutation) UpdateInvitation(ct context.Context, args struct {
	InvitationID graphql.ID
	Input        struct {
		SenderUserID   *graphql.ID
		ReceiverUserID *graphql.ID
		ExpireAt       *graphql.Time
		Status         *entity.InvitationStatus
	}
}) (Invitation, error) {
	panic("implement me")
}

func (m Mutation) DeleteInvitation(ct context.Context, args struct {
	InvitationID graphql.ID
}) (Invitation, error) {
	panic("implement me")
}

func (m Mutation) AcceptInvitation(ct context.Context, args struct {
	InvitationCode string
}) (Invitation, error) {
	panic("implement me")
}

func (m Mutation) DeclineInvitation(ct context.Context, args struct {
	InvitationCode string
}) (Invitation, error) {
	panic("implement me")
}
