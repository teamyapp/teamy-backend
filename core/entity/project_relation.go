package entity

type ProjectPhaseRelation struct {
	ProjectID uint64
	PhaseID   uint64
}

type ProjectStoryRelation struct {
	ProjectID uint64
	StoryID   uint64
}

type TeamProjectRelation struct {
	TeamID    uint64
	ProjectID uint64
}
