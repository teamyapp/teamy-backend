package realtime

type TeamActivity struct {
	teamID           uint64
	clientActivities map[uint64]*ClientActivity
}

func (t *TeamActivity) getClientActivity(clientID uint64) *ClientActivity {
	clientActivity, ok := t.clientActivities[clientID]
	if !ok {
		clientActivity = newClientActivity(clientID)
	}

	return clientActivity
}

func newTeamActivity(teamID uint64) *TeamActivity {
	teamActivity := TeamActivity{
		teamID:           teamID,
		clientActivities: map[uint64]*ClientActivity{},
	}

	return &teamActivity
}
