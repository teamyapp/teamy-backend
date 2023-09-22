package entity

type Importance string

const (
	BrandImpactImportance Importance = "BRAND_IMPACT"
	HighImpactImportance  Importance = "HIGH_IMPACT"
	MustHaveImportance    Importance = "MUST_HAVE"
	NiceToHaveImportance  Importance = "NICE_TO_HAVE"
)
