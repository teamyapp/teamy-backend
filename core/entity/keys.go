package entity

import "fmt"

func GetActivatorKey(id uint64) string {
	return fmt.Sprintf("Activator/%v", id)
}

func GetTeamAppInstallationKey(id uint64) string {
	return fmt.Sprintf("TeamAppInstallation/%v", id)
}

func GetAppPackageUploadSessionKey(appID, versionNumber uint64) string {
	return fmt.Sprintf("AppPackageUploadSession/%v:%v", appID, versionNumber)
}

func GetAppSecretKey(id uint64) string {
	return fmt.Sprintf("AppSecret/%v", id)
}

func GetAppTagRelationKey(appID, tagID uint64) string {
	return fmt.Sprintf("AppTagRelation/%v:%v", appID, tagID)
}

func GetAppVersionKey(appID, number uint64) string {
	return fmt.Sprintf("AppVersion/%v:%v", appID, number)
}

func GetAppVersionChangeKey(id uint64) string {
	return fmt.Sprintf("AppVersionChange/%v", id)
}

func GetAppVersionPriceKey(appID, versionNumber uint64, currency string) string {
	return fmt.Sprintf("AppVersionPrice/%v:%v:%v", appID, versionNumber, currency)
}

func GetAppKey(id uint64) string {
	return fmt.Sprintf("App/%v", id)
}

func GetClientKey(id uint64) string {
	return fmt.Sprintf("Client/%v", id)
}

func GetFileMetadataKey(id uint64) string {
	return fmt.Sprintf("FileMetadata/%v", id)
}

func GetGroupMemberRelationKey(groupID, memberID uint64) string {
	return fmt.Sprintf("GroupMemberRelation/%v:%v", groupID, memberID)
}

func GetGroupKey(id uint64) string {
	return fmt.Sprintf("Group/%v", id)
}

func GetInvitationKey(id uint64) string {
	return fmt.Sprintf("Invitation/%v", id)
}

func GetMessageKey(id uint64) string {
	return fmt.Sprintf("Message/%v", id)
}

func GetMoneyKey(currency string) string {
	return fmt.Sprintf("Money/%v", currency)
}

func GetPhaseStoryRelationKey(phaseID, storyID uint64) string {
	return fmt.Sprintf("PhaseStoryRelation/%v:%v", phaseID, storyID)
}

func GetPhaseKey(id uint64) string {
	return fmt.Sprintf("Phase/%v", id)
}

func GetProjectPhaseRelationKey(projectID, phaseID uint64) string {
	return fmt.Sprintf("ProjectPhaseRelation/%v:%v", projectID, phaseID)
}

func GetTeamProjectRelationKey(teamID, projectID uint64) string {
	return fmt.Sprintf("TeamProjectRelation/%v:%v", teamID, projectID)
}

func GetProjectKey(id uint64) string {
	return fmt.Sprintf("Project/%v", id)
}

func GetAppRolloutRelationKey(appID, rolloutID uint64) string {
	return fmt.Sprintf("AppRolloutRelation/%v:%v", appID, rolloutID)
}

func GetRolloutViewerKey(rolloutID, viewerID uint64) string {
	return fmt.Sprintf("RolloutViewer/%v:%v", rolloutID, viewerID)
}

func GetRolloutKey(id uint64) string {
	return fmt.Sprintf("Rollout/%v", id)
}

func GetSprintParticipantKey(sprintID, userID uint64) string {
	return fmt.Sprintf("SprintParticipant/%v:%v", sprintID, userID)
}

func GetSprintTaskRelationKey(sprintID, taskID uint64) string {
	return fmt.Sprintf("SprintTaskRelation/%v:%v", sprintID, taskID)
}

func GetSprintKey(id uint64) string {
	return fmt.Sprintf("Sprint/%v", id)
}

func GetStoryTaskRelationKey(storyID, taskID uint64) string {
	return fmt.Sprintf("StoryTaskRelation/%v:%v", storyID, taskID)
}

func GetStoryKey(id uint64) string {
	return fmt.Sprintf("Story/%v", id)
}

func GetTagKey(id uint64) string {
	return fmt.Sprintf("Tag/%v", id)
}

func GetTaskActivityKey(teamID, taskID uint64) string {
	return fmt.Sprintf("TaskActivity/%v:%v", teamID, taskID)
}

func GetTaskAwaitForRelationKey(awaitingTaskID, awaitForTaskID uint64) string {
	return fmt.Sprintf("TaskAwaitForRelation/%v:%v", awaitingTaskID, awaitForTaskID)
}

func GetTaskLinkKey(id uint64) string {
	return fmt.Sprintf("TaskLink/%v", id)
}

func GetTaskKey(id uint64) string {
	return fmt.Sprintf("Task/%v", id)
}

func GetTeamFileUploadSessionKey(teamID, sessionType uint64) string {
	return fmt.Sprintf("TeamFileUploadSession/%v:%v", teamID, sessionType)
}

func GetTeamMemberGroupKey(id uint64) string {
	return fmt.Sprintf("TeamMemberGroup/%v", id)
}

func GetTeamMemberKey(teamID, userID uint64) string {
	return fmt.Sprintf("TeamMember/%v:%v", teamID, userID)
}

func GetTeamKey(id uint64) string {
	return fmt.Sprintf("Team/%v", id)
}

func GetTasksForTeamKey(teamID uint64) string {
	return fmt.Sprintf("TasksForTeam/%v", teamID)
}

func GetUserFileUploadSessionKey(userID, sessionType uint64) string {
	return fmt.Sprintf("UserFileUploadSession/%v:%v", userID, sessionType)
}

func GetUserKey(id uint64) string {
	return fmt.Sprintf("User/%v", id)
}

func GetVersionSelectorVersionRelationKey(versionSelectorID, versionNumber uint64) string {
	return fmt.Sprintf("VersionSelectorVersionRelation/%v:%v", versionSelectorID, versionNumber)
}

func GetVersionSelectorKey(id uint64) string {
	return fmt.Sprintf("VersionSelector/%v", id)
}
