package gql

type SortOrder string

const (
	ASC  SortOrder = "ASC"
	DESC SortOrder = "DESC"
)

type TaskSortField string

const (
	TaskSortFieldPriority TaskSortField = "PRIORITY"
)
