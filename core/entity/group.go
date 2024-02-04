package entity

import (
	"time"
)

type GroupType string

const (
	GroupTypeStatic GroupType = "STATIC"
	GroupTypeFilter GroupType = "FILTER"
)

type GroupMemberType string

const (
	GroupMemberTypeUser GroupMemberType = "USER"
	GroupMemberTypeTeam GroupMemberType = "TEAM"
)

type Group struct {
	ID         uint64
	Type       GroupType
	MemberType GroupMemberType
	Name       string
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type GroupUnion struct {
	Type        GroupType
	MemberType  GroupMemberType
	StaticGroup StaticGroup
	FilterGroup FilterGroup
}

type StaticGroup struct {
	Group
}

type FilterGroup struct {
	Group
	Filter string
	Count  int
}
