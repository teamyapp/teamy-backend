package realtime

type CollectionType string

const (
	TaskCollectionType                 CollectionType = "Task"
	TaskLinkCollectionType             CollectionType = "TaskLink"
	InvitationCollectionType           CollectionType = "Invitation"
	MessageCollectionType              CollectionType = "Message"
	TeamCollectionType                 CollectionType = "Team"
	TeamMemberGroupCollectionType      CollectionType = "TeamMemberGroup"
	UserCollectionType                 CollectionType = "User"
	ThreadCollectionType               CollectionType = "Thread"
	TeamMemberCollectionType           CollectionType = "TeamMember"
	TaskAwaitForRelationCollectionType CollectionType = "TaskAwaitForRelation"
	SprintCollectionType               CollectionType = "Sprint"
	SprintTaskRelationCollectionType   CollectionType = "SprintTaskRelation"
	ClientCollectionType               CollectionType = "Client"
	TaskActivityCollectionType         CollectionType = "TaskActivity"
	SprintParticipantCollectionType    CollectionType = "SprintParticipant"
)
