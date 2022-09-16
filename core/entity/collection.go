package entity

type CollectionType string

const (
	TaskCollectionType                 CollectionType = "Task"
	InvitationCollectionType           CollectionType = "Invitation"
	MessageCollectionType              CollectionType = "Message"
	TeamCollectionType                 CollectionType = "Team"
	UserCollectionType                 CollectionType = "User"
	ThreadCollectionType               CollectionType = "Thread"
	TeamMemberCollectionType           CollectionType = "TeamMember"
	TaskAwaitForRelationCollectionType CollectionType = "TaskAwaitForRelation"
	SprintTaskRelationCollectionType   CollectionType = "SprintTaskRelation"
	ClientCollectionType               CollectionType = "Client"
	TaskActivityCollectionType         CollectionType = "TaskActivity"
)
