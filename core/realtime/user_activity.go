package realtime

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserActivity struct {
	userID         uint64
	state          entity.UserStatus
	teamActivities map[uint64]*TeamActivity
}

func (u *UserActivity) getTeamActivity(teamID uint64) *TeamActivity {
	teamActivity, ok := u.teamActivities[teamID]

	if !ok {
		teamActivity = u.newTeamActivity(teamID)
	}

	return teamActivity
}

func (u *UserActivity) newTeamActivity(teamID uint64) *TeamActivity {
	teamActivity := newTeamActivity(teamID)
	u.teamActivities[teamID] = teamActivity
	return teamActivity
}

func newUserActivity(userID uint64) *UserActivity {
	userActivity := &UserActivity{
		userID:         userID,
		state:          entity.UserStatusInitialized,
		teamActivities: map[uint64]*TeamActivity{},
	}

	return userActivity
}
