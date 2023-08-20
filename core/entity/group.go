package entity

import (
	"time"
)

type GroupType string

const (
	GroupTypeStatic GroupType = "STATIC"
	GroupTypeFilter GroupType = "FILTER"
)

type Group struct {
	ID        uint64
	Type      GroupType
	Name      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type FilterGroup struct {
	Group
	Filter string
	Count  int
}
