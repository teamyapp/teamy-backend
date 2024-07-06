package service

type SortOrder string

const (
	SortOrderAsc  SortOrder = "ASC"
	SortOrderDesc SortOrder = "DESC"
)

type TaskSortField string

const (
	TaskSortFieldPriority TaskSortField = "PRIORITY"
)
