package entity

type VersionSelectorType string

const (
	VersionSelectorTypeStatic     VersionSelectorType = "STATIC"
	VersionSelectorTypeExperiment VersionSelectorType = "EXPERIMENT"
)

type VersionSelector struct {
	ID   uint64
	Type VersionSelectorType
}
