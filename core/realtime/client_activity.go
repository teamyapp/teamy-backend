package realtime

type TaskAction struct {
	dragging bool
	taskID   uint64
}

type ClientActivity struct {
	clientID   uint64
	taskAction *TaskAction
}

func newClientActivity(clientID uint64) *ClientActivity {
	clientActivity := ClientActivity{
		clientID:   clientID,
		taskAction: &TaskAction{},
	}

	return &clientActivity
}
