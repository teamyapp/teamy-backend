package entity

type VersionSelectorType string

const (
	VersionSelectorTypeStatic     VersionSelectorType = "STATIC"
	VersionSelectorTypeExperiment VersionSelectorType = "EXPERIMENT"
)

type VersionSelectorUnion struct {
	Type                      VersionSelectorType
	StaticVersionSelector     StaticVersionSelector
	ExperimentVersionSelector ExperimentVersionSelector
}

type VersionSelector struct {
	ID   uint64
	Type VersionSelectorType
}

type StaticVersionSelector struct {
	VersionSelector
	VersionNumber int
}

type ExperimentVersionSelector struct {
	VersionSelector
	VersionNumbers []int
}
