package datastore

import (
	"fmt"

	"github.com/teamyapp/teamy-backend/app/entity"
)

func (d DataStore) CreateTeam(userID uint64, t entity.Team) (entity.Team, error) {
	t.ID = d.newID(Team)
	d.data.Teams = append(d.data.Teams, t)
	return t, d.persister.Write(d.data)
}

func (d DataStore) FilterTeams(filter func(entity.Team) bool) (ts []entity.Team) {
	for _, t := range d.data.Teams {
		if filter(t) {
			if t.NeedAttentionTasks == nil {
				t.NeedAttentionTasks = make(map[uint64]uint64)
			}
			ts = append(ts, t)
		}
	}
	return
}

func (d DataStore) UpdateTeam(teamID uint64, apply func(entity.Team) entity.Team) (entity.Team, error) {
	for i, team := range d.data.Teams {
		if team.ID == teamID {
			if team.NeedAttentionTasks == nil {
				team.NeedAttentionTasks = make(map[uint64]uint64)
			}
			newTeam := apply(team)
			d.data.Teams[i] = newTeam
			return newTeam, d.persister.Write(d.data)
		}
	}
	return entity.Team{}, fmt.Errorf("team %v is not found", teamID)
}
