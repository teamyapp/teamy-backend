package dao

type DataSource struct {
	userDao                 User
	teamDao                 Team
	teamMemberDao           TeamMember
	invitationDao           Invitation
	taskDao                 Task
	threadDao               Thread
	messageDao              Message
	taskAwaitForRelationDao TaskAwaitForRelation
}

func (d DataSource) UserDao() User {
	return d.userDao
}

func (d DataSource) TeamDao() Team {
	return d.teamDao
}

func (d DataSource) TeamMemberDao() TeamMember {
	return d.teamMemberDao
}

func (d DataSource) InvitationDao() Invitation {
	return d.invitationDao
}

func (d DataSource) TaskDao() Task {
	return d.taskDao
}

func (d DataSource) ThreadDao() Thread {
	return d.threadDao
}

func (d DataSource) MessageDao() Message {
	return d.messageDao
}

func (d DataSource) TaskAwaitForRelationDao() TaskAwaitForRelation {
	return d.taskAwaitForRelationDao
}

func NewDataSource(
	userDao User,
	teamDao Team,
	teamMemberDao TeamMember,
	invitationDao Invitation,
	taskDao Task,
	threadDao Thread,
	messageDao Message,
	taskAwaitForRelationDao TaskAwaitForRelation,
) DataSource {
	return DataSource{
		userDao:                 userDao,
		teamDao:                 teamDao,
		teamMemberDao:           teamMemberDao,
		invitationDao:           invitationDao,
		taskDao:                 taskDao,
		threadDao:               threadDao,
		messageDao:              messageDao,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
	}
}
